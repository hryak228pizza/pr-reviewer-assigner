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

type TeamUseCase struct {
	repo       repository.TeamRepository
	transactor repository.Transactor
}

func NewTeamUseCase(repo repository.TeamRepository, transactor repository.Transactor) *TeamUseCase {
	return &TeamUseCase{
		repo:       repo,
		transactor: transactor,
	}
}

func (uc *TeamUseCase) CreateTeamWithUsers(ctx context.Context, team *entity.Team, users []*entity.User) error {

	err := uc.transactor.Do(ctx, func(txCtx context.Context) error {
		if err := uc.repo.Create(txCtx, team, users); err != nil {
			return fmt.Errorf("TeamUseCase.CreateTeamWithUsers failed repo call: %w", err)
		}
		return nil 
	})

	if err != nil {
		if errors.Is(err, entity.ErrTeamExists) {
			return entity.ErrTeamExists
		}
	}

	return err
}
