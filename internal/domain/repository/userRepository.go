package repository

import (
	"context"
	"github.com/hryak228pizza/pr-reviewer-assigner/internal/domain/entity"
)

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
	GetActiveCandidatesByTeam(ctx context.Context, teamName string, excludeUserID string) ([]*entity.User, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) (*entity.User, error)
}
