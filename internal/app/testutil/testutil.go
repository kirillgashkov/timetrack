package testutil

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewTestPool() *pgxpool.Pool {
	dsn := os.Getenv("TEST_APP_DATABASE_DSN")
	if dsn == "" {
		panic("TEST_APP_DATABASE_DSN is not set")
	}

	db, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		panic(err)
	}

	return db
}
