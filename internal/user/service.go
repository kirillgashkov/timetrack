package user

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"strings"

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

type Filter struct {
	PassportNumber *string
	Surname        *string
	Name           *string
	Patronymic     *sql.NullString
	Address        *string
}

type Update struct {
	PassportNumber *string
	Surname        *string
	Name           *string
	Patronymic     *sql.NullString
	Address        *string
}

type Service interface {
	Create(ctx context.Context, passportNumber string) (*User, error)
	Get(ctx context.Context, id int) (*User, error)
	List(ctx context.Context, filter *Filter, offset, limit int) ([]User, error)
	Update(ctx context.Context, id int, update *Update) (*User, error)
	Delete(ctx context.Context, id int) (*User, error)
}

type ServiceImpl struct {
	db                *pgxpool.Pool
	peopleInfoService PeopleInfoService
}

func NewServiceImpl(db *pgxpool.Pool, peopleInfoService PeopleInfoService) *ServiceImpl {
	return &ServiceImpl{db: db, peopleInfoService: peopleInfoService}
}

var (
	ErrAlreadyExists         = errors.New("user already exists")
	ErrNotFound              = errors.New("user not found")
	ErrInvalidPassportNumber = errors.New("invalid passport number")
)

func (s *ServiceImpl) Create(ctx context.Context, passportNumber string) (*User, error) {
	series, number, err := parsePassportNumber(passportNumber)
	if err != nil {
		return nil, errors.Join(ErrInvalidPassportNumber, err)
	}

	info, err := s.peopleInfoService.Get(ctx, series, number)
	if err != nil {
		return nil, err
	}

	q := `
		INSERT INTO users (passport_number, surname, name, patronymic, address)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, passport_number, surname, name, patronymic, address
	`
	args := []any{
		passportNumber,
		info.Surname,
		info.Name,
		info.Patronymic,
		info.Address,
	}
	return s.queryOne(ctx, q, args...)
}

func (s *ServiceImpl) Get(ctx context.Context, id int) (*User, error) {
	q := `
		SELECT id, passport_number, surname, name, patronymic, address
		FROM users
		WHERE id = $1
	`
	return s.queryOne(ctx, q, id)
}

func (s *ServiceImpl) List(ctx context.Context, filter *Filter, offset, limit int) ([]User, error) {
	q, args := buildSelectQuery(filter, limit, offset)
	return s.queryAll(ctx, q, args...)
}

func (s *ServiceImpl) Update(ctx context.Context, id int, update *Update) (*User, error) {
	q := `
		UPDATE users
		SET passport_number = COALESCE($2, passport_number),
			surname = COALESCE($3, surname),
			name = COALESCE($4, name),
			patronymic = CASE WHEN $6 THEN $5 ELSE patronymic END,
			address = COALESCE($7, address)
		WHERE id = $1
		RETURNING id, passport_number, surname, name, patronymic, address
	`
	args := []any{
		id,
		update.PassportNumber,
		update.Surname,
		update.Name,
		update.Patronymic,
		update.Patronymic != nil,
		update.Address,
	}
	return s.queryOne(ctx, q, args...)
}

func (s *ServiceImpl) Delete(ctx context.Context, id int) (*User, error) {
	q := `
		DELETE FROM users
		WHERE id = $1
		RETURNING id, passport_number, surname, name, patronymic, address
	`
	return s.queryOne(ctx, q, id)
}

func (s *ServiceImpl) queryAll(ctx context.Context, query string, args ...any) ([]User, error) {
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Join(errors.New("failed to select users"), err)
	}
	defer rows.Close()

	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[User])
	if err != nil {
		return nil, errors.Join(errors.New("failed to collect users"), err)
	}
	return users, nil
}

func (s *ServiceImpl) queryOne(ctx context.Context, query string, args ...any) (*User, error) {
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, errors.Join(errors.New("failed to select user"), err)
	}
	defer rows.Close()

	user, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[User])
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return nil, errors.Join(ErrAlreadyExists, err)
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.Join(ErrNotFound, ErrNotFound)
		}
		return nil, errors.Join(errors.New("failed to collect user"), err)
	}
	return &user, nil
}

// buildSelectQuery builds a SELECT query with WHERE conditions based on the
// provided filter. Filters utilize the similarity operator % for string
// comparison (pg_trgm).
func buildSelectQuery(filter *Filter, limit, offset int) (string, []any) {
	baseQuery := `
		SELECT id, passport_number, surname, name, patronymic, address
		FROM users
	`
	whereConditions := make([]string, 0)
	args := make([]any, 0)
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
		if filter.Patronymic.Valid {
			whereConditions = append(whereConditions, `patronymic % $`+itoa(argIndex))
			args = append(args, *filter.Patronymic)
			argIndex++
		} else {
			whereConditions = append(whereConditions, `patronymic IS NULL`)
		}
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

// parsePassportNumber parses a passport number string into a series and a
// number.
func parsePassportNumber(passportNumber string) (int, int, error) {
	parts := strings.Split(passportNumber, " ")
	if len(parts) != 2 {
		return 0, 0, errors.New("invalid passport number")
	}

	series, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, errors.Join(errors.New("failed to parse passport series"), err)
	}
	if series < 0 || series > 9999 {
		return 0, 0, errors.New("passport series out of range")
	}

	number, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, errors.Join(errors.New("failed to parse passport number"), err)
	}
	if number < 0 || number > 999999 {
		return 0, 0, errors.New("passport number out of range")
	}

	return series, number, nil
}

func itoa(i int) string {
	return strconv.Itoa(i)
}
