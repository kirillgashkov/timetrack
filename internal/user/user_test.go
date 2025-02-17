package user

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kirillgashkov/timetrack/internal/app/database"
	"github.com/kirillgashkov/timetrack/internal/app/testutil"
)

var (
	poolDB *pgxpool.Pool
)

func TestMain(m *testing.M) {
	exitCode := func() int {
		poolDB = testutil.NewTestPool()
		return m.Run()
	}()
	os.Exit(exitCode)
}

func TestPostUsers(t *testing.T) {
	txDB := beginTx(poolDB)
	defer rollbackTx(txDB)

	handler := newTestHandler(txDB)

	t.Run("ok", func(t *testing.T) {
		resp := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/", bytes.NewBufferString(`{"passportNumber":"1234 567890"}`))

		handler.PostUsers(resp, req)

		got := resp.Code
		want := http.StatusOK

		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

func TestGetUsers(t *testing.T) {}

func TestGetUsersCurrent(t *testing.T) {}

func TestGetUsersId(t *testing.T) {}

func TestPatchUsersId(t *testing.T) {}

func TestDeleteUsersId(t *testing.T) {}

func beginTx(db database.DB) pgx.Tx {
	tx, err := db.Begin(context.TODO())
	if err != nil {
		panic(err)
	}
	return tx
}

func rollbackTx(tx pgx.Tx) {
	if txErr := tx.Rollback(context.TODO()); txErr != nil && !errors.Is(txErr, pgx.ErrTxClosed) {
		panic(txErr)
	}
}

func newTestHandler(db database.DB) *Handler {
	service := newTestServiceImpl(db)
	return NewHandler(service)
}

func newTestServiceImpl(db database.DB) *ServiceImpl {
	peopleInfoService := newTestPeopleInfoServiceImpl()
	return NewServiceImpl(db, peopleInfoService)
}

func newTestPeopleInfoServiceImpl() *PeopleInfoServiceImpl {
	serverURL := os.Getenv("TEST_APP_PEOPLE_INFO_SERVER_URL")
	if serverURL == "" {
		panic("TEST_APP_PEOPLE_INFO_SERVER_URL is not set")
	}
	httpClient := &http.Client{Timeout: 5 * time.Second}

	service, err := NewPeopleInfoServiceImpl(serverURL, httpClient)
	if err != nil {
		panic(err)
	}

	return service
}
