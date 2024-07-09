package tracking

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/jackc/pgx/v5"
	"github.com/kirillgashkov/timetrack/db/timetrackdb"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

type UserID int
type TaskID int

var (
	ErrAlreadyStarted = errors.New("task already started")
	ErrNotStarted     = errors.New("task not started")
)

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) StartTask(ctx context.Context, taskID TaskID, userID UserID) error {
	_, err := s.db.Exec(
		ctx,
		`INSERT INTO works (started_at, task_id, user_id, status) VALUES (now(), $1, $2, $3)`,
		taskID,
		userID,
		timetrackdb.WorkStatusStarted,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return ErrAlreadyStarted
		}
		return errors.Join(errors.New("failed to insert work"), err)
	}
	return nil
}

func (s *Service) StopTask(ctx context.Context, taskID TaskID, userID UserID) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return errors.Join(errors.New("failed to start transaction"), err)
	}
	defer func(tx pgx.Tx) {
		if txErr := tx.Rollback(ctx); txErr != nil && !errors.Is(txErr, pgx.ErrTxClosed) {
			slog.Error("failed to rollback transaction", "error", txErr)
		}
	}(tx)

	tag, err := s.db.Exec(
		ctx,
		`UPDATE works SET stopped_at = now(), status = $1 WHERE task_id = $2 AND user_id = $3 AND status = $4`,
		timetrackdb.WorkStatusStopped,
		taskID,
		userID,
		timetrackdb.WorkStatusStarted,
	)
	if err != nil {
		return errors.Join(errors.New("failed to update work"), err)
	}
	if tag.RowsAffected() != 1 {
		if tag.RowsAffected() == 0 {
			return ErrNotStarted
		}
		return fmt.Errorf("update work affected unexpected number of rows: %d", tag.RowsAffected())
	}

	return tx.Commit(ctx)
}
