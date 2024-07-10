package auth

import (
	"context"
	"errors"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	db *pgxpool.Pool
}

type PasswordGrant struct {
	Username string
	Password string
}

type Token struct {
	AccessToken string
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

func (s *Service) Auth(ctx context.Context, g *PasswordGrant) (*Token, error) {
	rows, err := s.db.Query(ctx, `SELECT id FROM users WHERE passport_number = $1`, g.Username)
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

	// Pseudo-authentication that checks if the username and password are the
	// same.
	if g.Username != g.Password {
		return nil, ErrInvalidCredentials
	}

	// Pseudo-token generation that uses the user ID as the access token.
	return &Token{AccessToken: strconv.Itoa(id)}, nil
}
