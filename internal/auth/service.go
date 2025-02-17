package auth

import (
	"context"
	"errors"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/kirillgashkov/timetrack/internal/app/database"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidAccessToken = errors.New("invalid access token")
)

type PasswordGrant struct {
	Username string
	Password string
}

type Token struct {
	AccessToken string
}

type User struct {
	ID int
}

type Service interface {
	Authorize(ctx context.Context, g *PasswordGrant) (*Token, error)
	UserFromAccessToken(accessToken string) (*User, error)
}

type ServiceImpl struct {
	db database.DB
}

func NewServiceImpl(db database.DB) *ServiceImpl {
	return &ServiceImpl{db: db}
}

func (s *ServiceImpl) Authorize(ctx context.Context, g *PasswordGrant) (*Token, error) {
	q := `SELECT id FROM users WHERE passport_number = $1`
	rows, err := s.db.Query(ctx, q, g.Username)
	if err != nil {
		return nil, errors.Join(errors.New("failed to select user"), err)
	}
	defer rows.Close()

	id, err := pgx.CollectExactlyOneRow(rows, pgx.RowTo[int])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, errors.Join(errors.New("failed to collect user"), err)
	}

	// Pseudo-authentication that uses the username as the password.
	if g.Username != g.Password {
		return nil, ErrInvalidCredentials
	}

	// Pseudo-token generation that uses the user ID as the access token.
	return &Token{AccessToken: strconv.Itoa(id)}, nil
}

func (s *ServiceImpl) UserFromAccessToken(accessToken string) (*User, error) {
	id, err := strconv.Atoi(accessToken)
	if err != nil {
		return nil, errors.Join(ErrInvalidAccessToken, err)
	}
	return &User{ID: id}, nil
}
