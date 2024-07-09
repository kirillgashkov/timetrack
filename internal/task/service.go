package task

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
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
		ctx,
		`INSERT INTO tasks (description) VALUES ($1) RETURNING id, description`,
		create.Description,
	)
	if err != nil {
		return nil, errors.Join(errors.New("failed to insert task"), err)
	}
	defer rows.Close()

	task, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Task])
	if err != nil {
		return nil, errors.Join(errors.New("failed to collect task"), err)
	}
	return &task, nil
}

func (s *Service) Get(ctx context.Context, id int) (*Task, error) {
	rows, err := s.db.Query(
		ctx,
		`SELECT id, description FROM tasks WHERE id = $1`,
		id,
	)
	if err != nil {
		return nil, errors.Join(errors.New("failed to select task"), err)
	}
	defer rows.Close()

	task, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Task])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, errors.Join(errors.New("failed to collect task"), err)
	}
	return &task, nil
}

func (s *Service) List(ctx context.Context, offset, limit int) ([]Task, error) {
	rows, err := s.db.Query(
		ctx,
		`SELECT id, description FROM tasks ORDER BY id OFFSET $1 LIMIT $2`,
		offset,
		limit,
	)
	if err != nil {
		return nil, errors.Join(errors.New("failed to select tasks"), err)
	}
	defer rows.Close()

	tasks, err := pgx.CollectRows(rows, pgx.RowToStructByName[Task])
	if err != nil {
		return nil, errors.Join(errors.New("failed to collect tasks"), err)
	}
	return tasks, nil
}

func (s *Service) Update(ctx context.Context, id int, update *Update) (*Task, error) {
	rows, err := s.db.Query(
		ctx,
		`UPDATE tasks SET description = coalesce($1, description) WHERE id = $2 RETURNING id, description`,
		update.Description,
		id,
	)
	if err != nil {
		return nil, errors.Join(errors.New("failed to update task"), err)
	}
	defer rows.Close()

	task, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Task])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, errors.Join(errors.New("failed to collect task"), err)
	}
	return &task, nil
}

func (s *Service) Delete(ctx context.Context, id int) (*Task, error) {
	rows, err := s.db.Query(
		ctx,
		`DELETE FROM tasks WHERE id = $1 RETURNING id, description`,
		id,
	)
	if err != nil {
		return nil, errors.Join(errors.New("failed to delete task"), err)
	}
	defer rows.Close()

	task, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Task])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, errors.Join(errors.New("failed to collect task"), err)
	}
	return &task, nil
}
