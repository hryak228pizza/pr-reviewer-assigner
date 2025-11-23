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

// TeamRepository manages team-related database operations
type TeamRepository struct {
	trm *postgres.TransactionManager
}

// NewTeamRepository creates new team repository instance
func NewTeamRepository(trm *postgres.TransactionManager) *TeamRepository {
	return &TeamRepository{trm: trm}
}

// check for interface implementation
var _ repository.TeamRepository = (*TeamRepository)(nil)

// Create creates a new team and optionally inserts members
func (r *TeamRepository) Create(ctx context.Context, team *entity.Team, users []*entity.User) error {
	queryer := r.trm.GetQueryer(ctx)

	teamQuery := `INSERT INTO teams (name) VALUES ($1) ON CONFLICT (name) DO NOTHING`

	// insert team if not exists
	teamTag, err := queryer.Exec(ctx, teamQuery, team.Name)
	if err != nil {
		return fmt.Errorf("TeamRepo.Create (team insert): %w", err)
	}

	if teamTag.RowsAffected() == 0 {
		return entity.ErrTeamExists
	}

	// insert initial team members using batch
	if len(users) > 0 {
		batch := &pgx.Batch{}
		userQuery := `INSERT INTO users (id, username, team_name, is_active) VALUES ($1, $2, $3, $4)`

		for _, u := range users {
			batch.Queue(userQuery, u.ID, u.Username, team.Name, u.IsActive)
		}

		batchRes := queryer.SendBatch(ctx, batch)
		defer batchRes.Close()

		for range users {
			_, err := batchRes.Exec()
			if err != nil {
				return fmt.Errorf("TeamRepo.Create (user batch insert): %w", err)
			}
		}

		if err := batchRes.Close(); err != nil {
			return fmt.Errorf("TeamRepo.Create (batch close): %w", err)
		}
	}

	return nil
}

// GetByName retrieves team details by name
func (r *TeamRepository) GetByName(ctx context.Context, name string) (*entity.Team, error) {
	queryer := r.trm.GetQueryer(ctx)

	const query = `SELECT name FROM teams WHERE name = $1`

	var team entity.Team

	err := queryer.QueryRow(ctx, query, name).Scan(&team.Name)

	if err == pgx.ErrNoRows {
		return nil, entity.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("TeamRepo.GetByName: %w", err)
	}

	return &team, nil
}
