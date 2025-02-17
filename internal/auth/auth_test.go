package auth

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kirillgashkov/timetrack/internal/app/testutil"
)

var (
	db *pgxpool.Pool
)

type ServiceMock struct {
	AuthorizeFunc           func(ctx context.Context, g *PasswordGrant) (*Token, error)
	UserFromAccessTokenFunc func(accessToken string) (*User, error)
}

func (s *ServiceMock) Authorize(ctx context.Context, g *PasswordGrant) (*Token, error) {
	return s.AuthorizeFunc(ctx, g)
}

func (s *ServiceMock) UserFromAccessToken(accessToken string) (*User, error) {
	return s.UserFromAccessTokenFunc(accessToken)
}

func TestMain(m *testing.M) {
	code := func() int {
		db = testutil.NewTestPool()
		defer db.Close()

		return m.Run()
	}()
	os.Exit(code)
}

func setupUser(t *testing.T, passportNumber string) int {
	query := `
		INSERT INTO users (passport_number, surname, name, patronymic, address)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	args := []interface{}{
		passportNumber,
		"Surname " + passportNumber,
		"Name " + passportNumber,
		"Patronymic " + passportNumber,
		"Address " + passportNumber,
	}

	var userID int
	err := db.QueryRow(context.Background(), query, args...).Scan(&userID)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}
	return userID
}

func teardownUser(t *testing.T, id int) {
	if _, err := db.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, id); err != nil {
		t.Fatalf("failed to delete user: %v", err)
	}
}
