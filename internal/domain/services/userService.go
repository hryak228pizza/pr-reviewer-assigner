package services

import (
	"context"
	"github.com/hryak228pizza/pr-reviewer-assigner/internal/domain/entity"
	"github.com/hryak228pizza/pr-reviewer-assigner/internal/domain/repository"
)

type UserService interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (*entity.User, error)
	GetReviews(ctx context.Context, userID string) ([]*entity.PullRequest, error)
}

type UserUseCase struct {
	userRepo repository.UserRepository
	prRepo   repository.PRRepository
}

func NewUserUseCase(userRepo repository.UserRepository, prRepo repository.PRRepository) *UserUseCase {
	return &UserUseCase{
		userRepo: userRepo,
		prRepo:   prRepo,
	}
}

func (uc *UserUseCase) SetIsActive(ctx context.Context, userID string, isActive bool) (*entity.User, error) {
	updatedUser, err := uc.userRepo.SetIsActive(ctx, userID, isActive)
	return updatedUser, err
}

func (uc *UserUseCase) GetReviews(ctx context.Context, userID string) ([]*entity.PullRequest, error) {
	_, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return uc.prRepo.GetReviewsByUserID(ctx, userID)
}
