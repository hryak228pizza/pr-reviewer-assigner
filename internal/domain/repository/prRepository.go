package repository

import (
	"context"
	"github.com/hryak228pizza/pr-reviewer-assigner/internal/domain/entity"
)

type PRRepository interface {
	Create(ctx context.Context, pr *entity.PullRequest) error
	GetByID(ctx context.Context, id string) (*entity.PullRequest, error)
	UpdateStatus(ctx context.Context, id string, status entity.PRStatus) (*entity.PullRequest, error)
	SetReviewers(ctx context.Context, prID string, reviewerIDs []string) error
	GetReviewsByUserID(ctx context.Context, userID string) ([]*entity.PullRequest, error)
	GetReviewersByPRID(ctx context.Context, prID string) ([]*entity.User, error)
}
