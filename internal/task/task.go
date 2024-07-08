package task

import (
	"context"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

type Task struct {
	ID          int
	Description string
}

type Create struct {
	Description string
}

type Filter struct {
	Description *string
}

type Update struct {
	Description *string
}

var (
	ErrAlreadyExists = errors.New("task already exists")
	ErrNotFound      = errors.New("task not found")
)

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) Create(ctx context.Context, create *Create) (*Task, error) {
	rows, err := s.db.Query(
		ctx, `INSERT INTO tasks (description) VALUES ($1) RETURNING id, description`,
		create.Description,
	)
	if err != nil {
		return nil, errors.Join(errors.New("failed to insert task"), err)
	}
	defer rows.Close()

	task, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Task])
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return nil, ErrAlreadyExists
		}
		return nil, errors.Join(errors.New("failed to collect task"), err)
	}
	return &task, nil
}
