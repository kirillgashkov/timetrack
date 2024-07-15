package tracking

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kirillgashkov/timetrack/db/timetrackdb"
	"github.com/kirillgashkov/timetrack/internal/app/database"
)

var (
	ErrAlreadyStarted = errors.New("task already started")
	ErrNotStarted     = errors.New("task not started")
)

type UserID int

type TaskID int

type Service interface {
	StartTask(ctx context.Context, taskID TaskID, userID UserID) error
	StopTask(ctx context.Context, taskID TaskID, userID UserID) error
}

type ServiceImpl struct {
	db database.DB
}

func NewServiceImpl(db database.DB) *ServiceImpl {
	return &ServiceImpl{db: db}
}

func (s *ServiceImpl) StartTask(ctx context.Context, taskID TaskID, userID UserID) error {
	q := `INSERT INTO works (started_at, task_id, user_id, status) VALUES (now(), $1, $2, $3)`
	_, err := s.db.Exec(
		ctx,
		q,
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

func (s *ServiceImpl) StopTask(ctx context.Context, taskID TaskID, userID UserID) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return errors.Join(errors.New("failed to start transaction"), err)
	}
	defer func(tx pgx.Tx) {
		if txErr := tx.Rollback(ctx); txErr != nil && !errors.Is(txErr, pgx.ErrTxClosed) {
			slog.Error("failed to rollback transaction", "error", txErr)
		}
	}(tx)

	q := `UPDATE works SET stopped_at = now(), status = $1 WHERE task_id = $2 AND user_id = $3 AND status = $4`
	tag, err := s.db.Exec(
		ctx,
		q,
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
