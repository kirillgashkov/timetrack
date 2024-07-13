package auth_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
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
