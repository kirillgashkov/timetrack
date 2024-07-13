package auth

import (
	"context"
	"errors"
	"strconv"
	"testing"
)

func TestAuthorize(t *testing.T) {
	service := NewServiceImpl(db)

	t.Run("valid user credentials", func(t *testing.T) {
		passportNumber := "0200 000000"
		id := setupUser(t, passportNumber)
		defer teardownUser(t, id)

		token, err := service.Authorize(
			context.Background(), &PasswordGrant{Username: passportNumber, Password: passportNumber},
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
			context.Background(), &PasswordGrant{Username: passportNumber, Password: "wrong"},
		)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials, got: %v", err)
		}
		if token != nil {
			t.Errorf("expected nil token, got: %v", token)
		}
	})

	t.Run("non-existent user", func(t *testing.T) {
		passportNumber := "0404 000000"

		token, err := service.Authorize(
			context.Background(), &PasswordGrant{Username: passportNumber, Password: passportNumber},
		)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials, got: %v", err)
		}
		if token != nil {
			t.Errorf("expected nil token, got: %v", token)
		}
	})
}

func TestUserFromAccessToken(t *testing.T) {
	service := NewServiceImpl(db)

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
		if !errors.Is(err, ErrInvalidAccessToken) {
			t.Errorf("expected ErrInvalidAccessToken, got: %v", err)
		}
		if user != nil {
			t.Errorf("expected nil user, got: %v", user)
		}
	})
}
