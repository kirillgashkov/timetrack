package auth_test

import (
	"context"
	"errors"
	"os"
	"strconv"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kirillgashkov/timetrack/internal/auth"
)

var (
	db *pgxpool.Pool
)

func TestMain(m *testing.M) {
	code := func() int {
		var err error
		db, err = pgxpool.New(context.Background(), os.Getenv("TEST_APP_DATABASE_DSN"))
		if err != nil {
			panic("failed to connect to the database: " + err.Error())
		}
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

func TestAuthorize(t *testing.T) {
	service := auth.NewService(db)

	t.Run("valid user credentials", func(t *testing.T) {
		passportNumber := "0200 000000"
		id := setupUser(t, passportNumber)
		defer teardownUser(t, id)

		token, err := service.Authorize(
			context.Background(), &auth.PasswordGrant{Username: passportNumber, Password: passportNumber},
		)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if token == nil {
			t.Fatalf("expected token, got nil")
		}
		if token.AccessToken != strconv.Itoa(id) {
			t.Errorf("expected access token %d, got %s", id, token.AccessToken)
		}
	})

	t.Run("invalid user credentials", func(t *testing.T) {
		passportNumber := "0401 000000"
		id := setupUser(t, passportNumber)
		defer teardownUser(t, id)

		token, err := service.Authorize(
			context.Background(), &auth.PasswordGrant{Username: passportNumber, Password: "wrong"},
		)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, auth.ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials, got: %v", err)
		}
		if token != nil {
			t.Errorf("expected nil token, got: %v", token)
		}
	})

	t.Run("non-existent user", func(t *testing.T) {
		passportNumber := "0404 000000"

		token, err := service.Authorize(
			context.Background(), &auth.PasswordGrant{Username: passportNumber, Password: passportNumber},
		)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, auth.ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials, got: %v", err)
		}
		if token != nil {
			t.Errorf("expected nil token, got: %v", token)
		}
	})
}

func TestUserFromAccessToken(t *testing.T) {
	service := auth.NewService(db)

	t.Run("valid access token", func(t *testing.T) {
		passportNumber := "0200 000000"
		id := setupUser(t, passportNumber)
		defer teardownUser(t, id)

		user, err := service.UserFromAccessToken(strconv.Itoa(id))
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if user == nil {
			t.Fatalf("expected user, got nil")
		}
		if user.ID != id {
			t.Errorf("expected user ID %d, got %d", id, user.ID)
		}
	})

	t.Run("invalid access token", func(t *testing.T) {
		user, err := service.UserFromAccessToken("not_a_number")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, auth.ErrInvalidAccessToken) {
			t.Errorf("expected ErrInvalidAccessToken, got: %v", err)
		}
		if user != nil {
			t.Errorf("expected nil user, got: %v", user)
		}
	})
}
