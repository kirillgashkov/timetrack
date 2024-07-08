package track

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/kirillgashkov/timetrack/db/timetrackdb"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

var (
	ErrAlreadyStarted = errors.New("task already started")
	ErrNotStarted     = errors.New("task not started")
)

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) StartTask(ctx context.Context, taskID int, userID int) error {
	return s.startStopTask(ctx, taskID, userID, timetrackdb.TrackTypeStart)
}

func (s *Service) StopTask(ctx context.Context, taskID int, userID int) error {
	return s.startStopTask(ctx, taskID, userID, timetrackdb.TrackTypeStop)
}

func (s *Service) startStopTask(ctx context.Context, taskID int, userID int, startStopTrackType string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return errors.Join(errors.New("failed to start transaction"), err)
	}
	defer func(tx pgx.Tx) {
		if txErr := tx.Rollback(ctx); txErr != nil && !errors.Is(txErr, pgx.ErrTxClosed) {
			slog.Error("failed to rollback transaction", "error", txErr)
		}
	}(tx)

	// Select current status and lock it to prevent concurrent status updates.

	statusRows, err := s.db.Query(
		ctx,
		`SELECT status FROM tasks_users WHERE task_id = $1 AND user_id = $2 FOR UPDATE`,
		taskID,
		userID,
	)
	if err != nil {
		return errors.Join(errors.New("failed to select user's task status"), err)
	}
	defer statusRows.Close()

	status, err := pgx.CollectExactlyOneRow(statusRows, pgx.RowTo[string])
	if err != nil {
		return errors.Join(errors.New("failed to collect user's task status"), err)
	}

	// Validate the status and track type.

	switch startStopTrackType {
	case timetrackdb.TrackTypeStart:
		switch status {
		case timetrackdb.TaskUserStatusActive:
			return ErrAlreadyStarted
		case timetrackdb.TaskUserStatusInactive:
			// Continue.
		default:
			return fmt.Errorf("unexpected task status: %s", status)
		}
	case timetrackdb.TrackTypeStop:
		switch status {
		case timetrackdb.TaskUserStatusActive:
			// Continue.
		case timetrackdb.TaskUserStatusInactive:
			return ErrNotStarted
		default:
			return fmt.Errorf("unexpected task status: %s", status)
		}
	default:
		return fmt.Errorf("unexpected track type: %s", startStopTrackType)
	}

	// Insert a new track record.

	_, err = s.db.Exec(
		ctx,
		`INSERT INTO tracks (type, task_id, user_id) VALUES ($1, $2, $3)`,
		timetrackdb.TrackTypeStart,
		taskID,
		userID,
	)
	if err != nil {
		return errors.Join(errors.New("failed to insert track"), err)
	}

	// Update the task status.

	_, err = s.db.Exec(
		ctx,
		`UPDATE tasks_users SET status = $1 WHERE task_id = $2 AND user_id = $3`,
		timetrackdb.TaskUserStatusActive,
		taskID,
		userID,
	)
	if err != nil {
		return errors.Join(errors.New("failed to update task status"), err)
	}

	return nil
}
