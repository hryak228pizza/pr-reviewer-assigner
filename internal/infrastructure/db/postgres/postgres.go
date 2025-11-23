// Package postgres manages database connections and transactions
package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMaxPoolSize  = 10
	defaultConnAttempts = 10
	defaultConnTimeout  = time.Second
)

type Postgres struct {
	maxPoolSize  int32
	connAttempts int
	connTimeout  time.Duration

	Pool *pgxpool.Pool
}

// New initializes new postgres instance
func New(ctx context.Context, url string) (*Postgres, error) {
	pg := &Postgres{
		maxPoolSize:  defaultMaxPoolSize,
		connAttempts: defaultConnAttempts,
		connTimeout:  defaultConnTimeout,
	}

	// parse connection string into config
	poolConfig, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("postgres - New - pgxpool.ParseConfig: %w", err)
	}

	poolConfig.MaxConns = int32(pg.maxPoolSize)

	// attempt to connect with retries
	for pg.connAttempts > 0 {
		pg.Pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
		if err == nil {
			// verify connection is alive
			err = pg.Pool.Ping(ctx)
			if err == nil {
				return pg, nil
			}
		}

		slog.Info("Postgres is trying to connect", "attempts_left", pg.connAttempts)
		// wait before next attempt
		time.Sleep(pg.connTimeout)
		pg.connAttempts--
	}

	return nil, fmt.Errorf("postgres - New - connAttempts == 0: %w", err)
}

// Close closes connection pool
func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}
