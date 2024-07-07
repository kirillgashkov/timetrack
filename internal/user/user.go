package user

import (
	"context"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID             int
	PassportNumber string `db:"passport_number"`
	Surname        string
	Name           string
	Patronymic     *string
	Address        string
}

type Service struct {
	db *pgxpool.Pool
}

var ErrAlreadyExists = errors.New("user already exists")

func (s *Service) Create(ctx context.Context, passportNumber string) (*User, error) {
	rows, err := s.db.Query(
		ctx,
		`
			INSERT INTO users (passport_number, surname, name, patronymic, address)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, passport_number, surname, name, patronymic, address
		`,
		passportNumber,
		"some surname",
		"some name",
		"some patronymic",
		"some address",
	)
	if err != nil {
		return nil, errors.Join(errors.New("failed to insert user"), err)
	}

	u, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[User])
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return nil, errors.Join(ErrAlreadyExists, err)
		}
		return nil, errors.Join(errors.New("failed to collect rows"), err)
	}
	return &u, nil
}
