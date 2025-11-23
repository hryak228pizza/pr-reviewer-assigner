package repository

import (
	"context"
)

type Transactor interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
