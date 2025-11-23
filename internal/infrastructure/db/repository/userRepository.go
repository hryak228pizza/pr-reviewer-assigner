// Package repository handles data persistence and retrieval
package repository

import (
	"context"
	"fmt"

	"github.com/hryak228pizza/pr-reviewer-assigner/internal/domain/entity"
	"github.com/hryak228pizza/pr-reviewer-assigner/internal/domain/repository"
	"github.com/hryak228pizza/pr-reviewer-assigner/internal/infrastructure/db/postgres"

	"github.com/jackc/pgx/v5"
)

// UserRepository manages user data access
type UserRepository struct {
	trm *postgres.TransactionManager
}

// NewUserRepository creates new user repository
func NewUserRepository(trm *postgres.TransactionManager) *UserRepository {
	return &UserRepository{trm: trm}
}

// check for interface implementation
var _ repository.UserRepository = (*UserRepository)(nil)

// GetByID fetches a user by unique identifier
func (r *UserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	queryer := r.trm.GetQueryer(ctx)

	const query = `SELECT id, username, team_name, is_active FROM users WHERE id = $1`

	user := &entity.User{}
	// execute query and scan result
	err := queryer.QueryRow(ctx, query, id).Scan(&user.ID, &user.Username, &user.TeamName, &user.IsActive)

	if err == pgx.ErrNoRows {
		return nil, entity.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("UserRepo.GetByID: %w", err)
	}

	return user, nil
}

// Create adds a new user to the database
func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	queryer := r.trm.GetQueryer(ctx)

	const query = `INSERT INTO users (id, username, team_name, is_active) VALUES ($1, $2, $3, $4)`

	_, err := queryer.Exec(ctx, query, user.ID, user.Username, user.TeamName, user.IsActive)
	if err != nil {
		return fmt.Errorf("UserRepo.Create: %w", err)
	}
	return nil
}

// GetActiveCandidatesByTeam retrieves potential reviewers from the same team
func (r *UserRepository) GetActiveCandidatesByTeam(ctx context.Context, teamName string, excludeUserID string) ([]*entity.User, error) {
	queryer := r.trm.GetQueryer(ctx)

	// filter by team, active status and exclude author
	const query = `
		SELECT id, username, team_name, is_active 
		FROM users 
		WHERE team_name = $1 AND is_active = TRUE AND id != $2`

	rows, err := queryer.Query(ctx, query, teamName, excludeUserID)
	if err != nil {
		return nil, fmt.Errorf("UserRepo.GetActiveCandidatesByTeam: %w", err)
	}
	defer rows.Close()

	users := make([]*entity.User, 0)
	for rows.Next() {
		user := &entity.User{}
		err := rows.Scan(&user.ID, &user.Username, &user.TeamName, &user.IsActive)
		if err != nil {
			return nil, fmt.Errorf("UserRepo.GetActiveCandidatesByTeam scan: %w", err)
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

// SetIsActive updates user's active status
func (r *UserRepository) SetIsActive(ctx context.Context, userID string, isActive bool) (*entity.User, error) {
	queryer := r.trm.GetQueryer(ctx)

	const query = `
		UPDATE users 
		SET is_active = $2 
		WHERE id = $1 
		RETURNING id, username, team_name, is_active`

	user := &entity.User{}
	// execute update and return modified user
	err := queryer.QueryRow(ctx, query, userID, isActive).Scan(&user.ID, &user.Username, &user.TeamName, &user.IsActive)

	if err == pgx.ErrNoRows {
		return nil, entity.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("UserRepo.SetIsActive: %w", err)
	}

	return user, nil
}
