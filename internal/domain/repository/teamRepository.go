// Package repository handles data persistence and retrieval
package repository

import (
	"context"

	"github.com/hryak228pizza/pr-reviewer-assigner/internal/domain/entity"
)

type TeamRepository interface {
	Create(ctx context.Context, team *entity.Team, users []*entity.User) error
	GetByName(ctx context.Context, name string) (*entity.Team, error)
}
