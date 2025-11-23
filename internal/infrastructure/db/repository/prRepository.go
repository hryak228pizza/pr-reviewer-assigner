// Package repository handles data persistence and retrieval
package repository

import (
	"context"
	"fmt"

	"github.com/hryak228pizza/pr-reviewer-assigner/internal/domain/entity"
	"github.com/hryak228pizza/pr-reviewer-assigner/internal/domain/repository"
	"github.com/hryak228pizza/pr-reviewer-assigner/internal/infrastructure/db/postgres"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// PRRepository manages pull request data persistence
type PRRepository struct {
	trm *postgres.TransactionManager
}

// NewPRRepository creates new pr repository instance
func NewPRRepository(trm *postgres.TransactionManager) *PRRepository {
	return &PRRepository{trm: trm}
}

// check for interface implementation
var _ repository.PRRepository = (*PRRepository)(nil)

// Create persists a new pull request to the database
func (r *PRRepository) Create(ctx context.Context, pr *entity.PullRequest) error {
	queryer := r.trm.GetQueryer(ctx)

	const prQuery = `
		INSERT INTO pull_requests (id, name, author_id, status, created_at) 
		VALUES ($1, $2, $3, $4, NOW())`

	// execute insert statement
	_, err := queryer.Exec(ctx, prQuery, pr.ID, pr.Name, pr.AuthorID, pr.Status)
	if err != nil {
		// check for duplicate primary key
		if err.Error() == "duplicate key value violates unique constraint \"pull_requests_pkey\"" {
			return entity.ErrPRExists
		}
		return fmt.Errorf("PRRepo.Create (pr insert): %w", err)
	}

	// add reviewers if any are specified
	if len(pr.Reviewers) > 0 {
		reviewerIDs := make([]string, len(pr.Reviewers))
		for i, rev := range pr.Reviewers {
			reviewerIDs[i] = rev.ID
		}
		if err := r.SetReviewers(ctx, pr.ID, reviewerIDs); err != nil {
			return fmt.Errorf("PRRepo.Create (set reviewers): %w", err)
		}
	}

	return nil
}

// GetByID retrieves a pull request by its unique identifier
func (r *PRRepository) GetByID(ctx context.Context, id string) (*entity.PullRequest, error) {
	queryer := r.trm.GetQueryer(ctx)

	const prQuery = `
		SELECT id, name, author_id, status, created_at, merged_at 
		FROM pull_requests 
		WHERE id = $1`

	pr := &entity.PullRequest{}
	var mergedAt pgtype.Timestamptz

	// scan basic pr details
	err := queryer.QueryRow(ctx, prQuery, id).Scan(
		&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &mergedAt)

	if err == pgx.ErrNoRows {
		return nil, entity.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("PRRepo.GetByID (pr fetch): %w", err)
	}

	// handle nullable merged_at timestamp
	if mergedAt.Valid {
		pr.MergedAt = &mergedAt.Time
	}

	const revQuery = `
		SELECT u.id, u.username, u.team_name, u.is_active 
		FROM pr_reviewers pr_rev
		JOIN users u ON pr_rev.reviewer_id = u.id
		WHERE pr_rev.pr_id = $1`

	// fetch associated reviewers
	rows, err := queryer.Query(ctx, revQuery, pr.ID)
	if err != nil {
		return nil, fmt.Errorf("PRRepo.GetByID (reviewers fetch): %w", err)
	}
	defer rows.Close()

	pr.Reviewers = make([]entity.User, 0)
	for rows.Next() {
		rev := entity.User{}
		if err := rows.Scan(&rev.ID, &rev.Username, &rev.TeamName, &rev.IsActive); err != nil {
			return nil, fmt.Errorf("PRRepo.GetByID (reviewers scan): %w", err)
		}
		pr.Reviewers = append(pr.Reviewers, rev)
	}

	return pr, rows.Err()
}

// UpdateStatus updates the status of an existing pull request
func (r *PRRepository) UpdateStatus(ctx context.Context, id string, status entity.PRStatus) (*entity.PullRequest, error) {
	queryer := r.trm.GetQueryer(ctx)

	setClause := "status = $2"
	args := []interface{}{id, status}

	// update merged_at timestamp if status changes to merged
	if status == entity.StatusMerged {
		setClause += ", merged_at = NOW()"
	}

	query := fmt.Sprintf(`
		UPDATE pull_requests 
		SET %s 
		WHERE id = $1 
		RETURNING id, name, author_id, status, created_at, merged_at`, setClause)

	pr := &entity.PullRequest{}
	var mergedAt pgtype.Timestamptz

	// execute update and return new state
	err := queryer.QueryRow(ctx, query, args...).Scan(
		&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &mergedAt)

	if err == pgx.ErrNoRows {
		return nil, entity.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("PRRepo.UpdateStatus: %w", err)
	}

	if mergedAt.Valid {
		pr.MergedAt = &mergedAt.Time
	}

	// refresh reviewers list
	reviewersPtrs, err := r.GetReviewersByPRID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("PRRepo.UpdateStatus: failed to load reviewers: %w", err)
	}

	reviewers := make([]entity.User, len(reviewersPtrs))
	for i, r := range reviewersPtrs {
		reviewers[i] = *r
	}

	pr.Reviewers = reviewers

	return pr, nil
}

// SetReviewers updates the list of reviewers for a specific pr
func (r *PRRepository) SetReviewers(ctx context.Context, prID string, reviewerIDs []string) error {
	queryer := r.trm.GetQueryer(ctx)

	// remove existing reviewers first
	const deleteQuery = `DELETE FROM pr_reviewers WHERE pr_id = $1`
	if _, err := queryer.Exec(ctx, deleteQuery, prID); err != nil {
		return fmt.Errorf("PRRepo.SetReviewers (delete): %w", err)
	}

	if len(reviewerIDs) == 0 {
		return nil
	}

	rows := make([][]interface{}, len(reviewerIDs))
	for i, userID := range reviewerIDs {
		rows[i] = []interface{}{prID, userID}
	}

	// bulk insert new reviewers
	_, err := queryer.CopyFrom(
		ctx,
		pgx.Identifier{"pr_reviewers"},
		[]string{"pr_id", "reviewer_id"},
		pgx.CopyFromRows(rows),
	)

	if err != nil {
		return fmt.Errorf("PRRepo.SetReviewers (copy from): %w", err)
	}

	return nil
}

// GetReviewsByUserID finds all pull requests assigned to a specific reviewer
func (r *PRRepository) GetReviewsByUserID(ctx context.Context, userID string) ([]*entity.PullRequest, error) {
	queryer := r.trm.GetQueryer(ctx)

	const query = `
		SELECT p.id, p.name, p.author_id, p.status 
		FROM pull_requests p
		JOIN pr_reviewers pr_rev ON p.id = pr_rev.pr_id
		WHERE pr_rev.reviewer_id = $1`

	// execute join query to find assignments
	rows, err := queryer.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("PRRepo.GetReviewsByUserID: %w", err)
	}
	defer rows.Close()

	prs := make([]*entity.PullRequest, 0)
	for rows.Next() {
		pr := &entity.PullRequest{}
		if err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status); err != nil {
			return nil, fmt.Errorf("PRRepo.GetReviewsByUserID scan: %w", err)
		}
		prs = append(prs, pr)
	}

	return prs, rows.Err()
}

// GetReviewersByPRID retrieves all users assigned to review a specific pr
func (r *PRRepository) GetReviewersByPRID(ctx context.Context, prID string) ([]*entity.User, error) {
	queryer := r.trm.GetQueryer(ctx)

	const query = `
        SELECT 
            u.id, u.username, u.team_name, u.is_active 
        FROM 
            pr_reviewers pr_rev 
        JOIN 
            users u ON pr_rev.reviewer_id = u.id 
        WHERE 
            pr_rev.pr_id = $1`

	rows, err := queryer.Query(ctx, query, prID)
	if err != nil {
		return nil, fmt.Errorf("PRRepo.GetReviewersByPRID (query): %w", err)
	}
	defer rows.Close()

	var reviewers []*entity.User

	for rows.Next() {
		user := &entity.User{}

		err := rows.Scan(&user.ID, &user.Username, &user.TeamName, &user.IsActive)
		if err != nil {
			return nil, fmt.Errorf("PRRepo.GetReviewersByPRID (scan): %w", err)
		}
		reviewers = append(reviewers, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("PRRepo.GetReviewersByPRID (rows error): %w", err)
	}

	return reviewers, nil
}
