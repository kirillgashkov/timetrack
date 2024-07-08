package database

import (
	"context"
	"errors"

	"github.com/kirillgashkov/timetrack/internal/app/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, cfg *config.DatabaseConfig) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(ctx, cfg.DSN)
	if err != nil {
		return nil, errors.Join(errors.New("failed to open database"), err)
	}
	if err = db.Ping(ctx); err != nil {
		return nil, errors.Join(errors.New("failed to ping database"), err)
	}
	return db, nil
}
