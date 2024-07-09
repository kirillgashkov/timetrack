package database

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/kirillgashkov/timetrack/internal/app/config"
)

func NewPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	pgxCfg, err := pgxpool.ParseConfig(cfg.Database.DSN)
	if err != nil {
		return nil, errors.Join(errors.New("failed to parse database config"), err)
	}

	switch cfg.Mode {
	case config.ModeDevelopment:
		pgxCfg.ConnConfig.Tracer = newTracer()
	case config.ModeProduction:
	default:
		panic("invalid mode")
	}

	db, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return nil, errors.Join(errors.New("failed to open database"), err)
	}
	if err = db.Ping(ctx); err != nil {
		return nil, errors.Join(errors.New("failed to ping database"), err)
	}
	return db, nil
}

func newTracer() *tracelog.TraceLog {
	loggerFunc := func(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]interface{}) {
		attrs := make([]slog.Attr, 0, len(data))
		for k, v := range data {
			attrs = append(attrs, slog.Any(k, v))
		}
		attrs = append(attrs, slog.String("source", "pgx"))

		var lvl slog.Level
		switch level {
		case tracelog.LogLevelTrace:
			lvl = slog.LevelDebug
		case tracelog.LogLevelDebug:
			lvl = slog.LevelDebug
		case tracelog.LogLevelInfo:
			lvl = slog.LevelInfo
		case tracelog.LogLevelWarn:
			lvl = slog.LevelWarn
		case tracelog.LogLevelError:
			lvl = slog.LevelError
		default:
			lvl = slog.LevelError
		}

		slog.LogAttrs(ctx, lvl, msg, attrs...)
	}
	return &tracelog.TraceLog{
		Logger:   tracelog.LoggerFunc(loggerFunc),
		LogLevel: tracelog.LogLevelDebug,
	}
}
