package user

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	PassportNumber string `db:"passport_number"`
	Surname        string
	Name           string
	Patronymic     *string
	Address        string
}

type Filter struct {
	PassportNumber  *string
	Surname         *string
	Name            *string
	Patronymic      *string
	PatronymicForce bool
	Address         *string
}

type Update struct {
	Surname         *string
	Name            *string
	Patronymic      *string
	PatronymicForce bool
	Address         *string
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
			RETURNING passport_number, surname, name, patronymic, address
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

func (s *Service) Get(ctx context.Context, passportNumber string) (*User, error) {
	rows, err := s.db.Query(
		ctx,
		`
			SELECT passport_number, surname, name, patronymic, address
			FROM users
			WHERE passport_number = $1
		`,
		passportNumber,
	)
	if err != nil {
		return nil, errors.Join(errors.New("failed to select user"), err)
	}

	u, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[User])
	if err != nil {
		return nil, errors.Join(errors.New("failed to collect rows"), err)
	}
	return &u, nil
}

func (s *Service) GetAll(ctx context.Context, filter Filter, limit, offset int) ([]User, error) {
	query, args := buildSelectQuery(filter, limit, offset)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Join(errors.New("failed to select users"), err)
	}

	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[User])
	if err != nil {
		return nil, errors.Join(errors.New("failed to collect rows"), err)
	}
	return users, nil
}

func (s *Service) Update(ctx context.Context, passportNumber string, update Update) (*User, error) {
	rows, err := s.db.Query(
		ctx,
		`
			UPDATE users
			SET surname = COALESCE($2, surname),
				name = COALESCE($3, name),
				patronymic = CASE $5 WHEN TRUE THEN $4 ELSE COALESCE($4, patronymic) END,
				address = COALESCE($6, address)
			WHERE passport_number = $1
			RETURNING passport_number, surname, name, patronymic, address
		`,
		passportNumber,
		update.Surname,
		update.Name,
		update.Patronymic,
		update.PatronymicForce,
		update.Address,
	)
	if err != nil {
		return nil, errors.Join(errors.New("failed to update user"), err)
	}

	u, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[User])
	if err != nil {
		return nil, errors.Join(errors.New("failed to collect rows"), err)
	}
	return &u, nil
}

func (s *Service) Delete(ctx context.Context, passportNumber string) (*User, error) {
	rows, err := s.db.Query(
		ctx,
		`
			DELETE FROM users
			WHERE passport_number = $1
			RETURNING passport_number, surname, name, patronymic, address
		`,
		passportNumber,
	)
	if err != nil {
		return nil, errors.Join(errors.New("failed to delete user"), err)
	}

	u, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[User])
	if err != nil {
		return nil, errors.Join(errors.New("failed to collect rows"), err)
	}
	return &u, nil
}

// buildSelectQuery builds a SELECT query with WHERE conditions based on the
// provided filter. Filters utilize the similarity operator % for string
// comparison (pg_trgm).
func buildSelectQuery(filter Filter, limit, offset int) (string, []any) {
	baseQuery := `
		SELECT id, passport_number, surname, name, patronymic, address
		FROM users
	`
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if filter.PassportNumber != nil {
		whereConditions = append(whereConditions, `passport_number % $`+itoa(argIndex))
		args = append(args, *filter.PassportNumber)
		argIndex++
	}

	if filter.Surname != nil {
		whereConditions = append(whereConditions, `surname % $`+itoa(argIndex))
		args = append(args, *filter.Surname)
		argIndex++
	}

	if filter.Name != nil {
		whereConditions = append(whereConditions, `name % $`+itoa(argIndex))
		args = append(args, *filter.Name)
		argIndex++
	}

	if filter.Patronymic != nil {
		whereConditions = append(whereConditions, `patronymic % $`+itoa(argIndex))
		args = append(args, *filter.Patronymic)
		argIndex++
	} else if filter.PatronymicForce {
		whereConditions = append(whereConditions, `patronymic IS NULL`)
	}

	if filter.Address != nil {
		whereConditions = append(whereConditions, `address % $`+itoa(argIndex))
		args = append(args, *filter.Address)
		argIndex++
	}

	if len(whereConditions) > 0 {
		baseQuery += "WHERE " + strings.Join(whereConditions, " AND ") + " "
	}

	baseQuery += "LIMIT $" + itoa(argIndex) + " OFFSET $" + itoa(argIndex+1)
	args = append(args, limit, offset)

	return baseQuery, args
}

func itoa(i int) string {
	return strconv.Itoa(i)
}
