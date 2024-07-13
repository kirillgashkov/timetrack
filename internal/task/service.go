package task

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound = errors.New("task not found")
)

type Service interface {
	Create(ctx context.Context, create *CreateTask) (*Task, error)
	Get(ctx context.Context, id int) (*Task, error)
	List(ctx context.Context, offset, limit int) ([]Task, error)
	Update(ctx context.Context, id int, update *UpdateTask) (*Task, error)
	Delete(ctx context.Context, id int) (*Task, error)
}

type Task struct {
	ID          int
	Description string
}

type CreateTask struct {
	Description string
}

type UpdateTask struct {
	Description *string
}

type ServiceImpl struct {
	db *pgxpool.Pool
}

func NewServiceImpl(db *pgxpool.Pool) *ServiceImpl {
	return &ServiceImpl{db: db}
}

func (s *ServiceImpl) Create(ctx context.Context, create *CreateTask) (*Task, error) {
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

func (s *ServiceImpl) Get(ctx context.Context, id int) (*Task, error) {
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

func (s *ServiceImpl) List(ctx context.Context, offset, limit int) ([]Task, error) {
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

func (s *ServiceImpl) Update(ctx context.Context, id int, update *UpdateTask) (*Task, error) {
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

func (s *ServiceImpl) Delete(ctx context.Context, id int) (*Task, error) {
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
