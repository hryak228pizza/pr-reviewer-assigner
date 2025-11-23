// Package services implements business logic and domain rules
package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/hryak228pizza/pr-reviewer-assigner/internal/domain/entity"
	"github.com/hryak228pizza/pr-reviewer-assigner/internal/domain/repository"
)

type TeamService interface {
	CreateTeamWithUsers(ctx context.Context, team *entity.Team, users []*entity.User) error
}

// TeamUseCase implements the TeamService interface
type TeamUseCase struct {
	repo       repository.TeamRepository
	transactor repository.Transactor
}

// NewTeamUseCase is the constructor for TeamUseCase
func NewTeamUseCase(repo repository.TeamRepository, transactor repository.Transactor) *TeamUseCase {
	return &TeamUseCase{
		repo:       repo,
		transactor: transactor,
	}
}

// CreateTeamWithUsers ensures team and users are created atomically or rolled back
func (uc *TeamUseCase) CreateTeamWithUsers(ctx context.Context, team *entity.Team, users []*entity.User) error {

	// start a transaction
	err := uc.transactor.Do(ctx, func(txCtx context.Context) error {
		if err := uc.repo.Create(txCtx, team, users); err != nil {
			// wrap the error to add context for tracing
			return fmt.Errorf("TeamUseCase.CreateTeamWithUsers failed repo call: %w", err)
		}
		return nil
	})

	if err != nil {
		if errors.Is(err, entity.ErrTeamExists) {
			// return error 409 conflict
			return entity.ErrTeamExists
		}
	}

	return err
}
