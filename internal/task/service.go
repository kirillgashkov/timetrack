package task

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/kirillgashkov/timetrack/internal/app/database"
)

var (
	ErrNotFound = errors.New("task not found")
)

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

type Service interface {
	Create(ctx context.Context, create *CreateTask) (*Task, error)
	Get(ctx context.Context, id int) (*Task, error)
	List(ctx context.Context, offset, limit int) ([]Task, error)
	Update(ctx context.Context, id int, update *UpdateTask) (*Task, error)
	Delete(ctx context.Context, id int) (*Task, error)
}

type ServiceImpl struct {
	db database.DB
}

func NewServiceImpl(db database.DB) *ServiceImpl {
	return &ServiceImpl{db: db}
}

func (s *ServiceImpl) Create(ctx context.Context, create *CreateTask) (*Task, error) {
	q := `INSERT INTO tasks (description) VALUES ($1) RETURNING id, description`
	return s.queryOne(ctx, q, create.Description)
}

func (s *ServiceImpl) Get(ctx context.Context, id int) (*Task, error) {
	q := `SELECT id, description FROM tasks WHERE id = $1`
	return s.queryOne(ctx, q, id)
}

func (s *ServiceImpl) List(ctx context.Context, offset, limit int) ([]Task, error) {
	q := `SELECT id, description FROM tasks ORDER BY id OFFSET $1 LIMIT $2`
	return s.queryAll(ctx, q, offset, limit)
}

func (s *ServiceImpl) Update(ctx context.Context, id int, update *UpdateTask) (*Task, error) {
	q := `UPDATE tasks SET description = coalesce($1, description) WHERE id = $2 RETURNING id, description`
	return s.queryOne(ctx, q, update.Description, id)
}

func (s *ServiceImpl) Delete(ctx context.Context, id int) (*Task, error) {
	q := `DELETE FROM tasks WHERE id = $1 RETURNING id, description`
	return s.queryOne(ctx, q, id)
}

func (s *ServiceImpl) queryAll(ctx context.Context, query string, args ...any) ([]Task, error) {
	rows, err := s.db.Query(ctx, query, args...)
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

func (s *ServiceImpl) queryOne(ctx context.Context, query string, args ...any) (*Task, error) {
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Join(errors.New("failed to select task"), err)
	}
	defer rows.Close()

	task, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Task])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.Join(ErrNotFound, ErrNotFound)
		}
		return nil, errors.Join(errors.New("failed to collect task"), err)
	}
	return &task, nil
}
