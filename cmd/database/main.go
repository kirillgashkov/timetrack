package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/kirillgashkov/assignment-timetrack/db/timetrackdb"
	"github.com/kirillgashkov/assignment-timetrack/internal/config"
	"github.com/kirillgashkov/assignment-timetrack/internal/logging"
)

func main() {
	if err := mainErr(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func mainErr() error {
	ctx := context.Background()

	cfg, err := config.New()
	if err != nil {
		return errors.Join(errors.New("failed to create config"), err)
	}

	logger := logging.NewLogger(cfg)
	slog.SetDefault(logger)

	db, err := newDB(ctx, cfg.DSN)
	if err != nil {
		return errors.Join(errors.New("failed to create database"), err)
	}
	defer shouldClose(db)

	if err = migrateDB(db); err != nil {
		return errors.Join(errors.New("failed to migrate database"), err)
	}
	return nil
}

func newDB(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, errors.Join(errors.New("failed to open database"), err)
	}
	if err = db.PingContext(ctx); err != nil {
		return nil, errors.Join(errors.New("failed to ping database"), err)
	}
	return db, nil
}

func migrateDB(db *sql.DB) error {
	sourceDriver, err := iofs.New(timetrackdb.Migrations(), "")
	if err != nil {
		return errors.Join(errors.New("failed to create migrate source driver"), err)
	}

	databaseDriver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return errors.Join(errors.New("failed to create migrate database driver"), err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", databaseDriver)
	if err != nil {
		return errors.Join(errors.New("failed to create migrate"), err)
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return errors.Join(errors.New("failed to migrate database"), err)
	}
	return nil
}

func shouldClose(c io.Closer) {
	if err := c.Close(); err != nil {
		slog.Error("failed to close resource", "error", err)
	}
}
