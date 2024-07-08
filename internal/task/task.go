package task

import (
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

var (
	ErrAlreadyExists = errors.New("task already exists")
	ErrNotFound      = errors.New("task not found")
)

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}
