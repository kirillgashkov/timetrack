package api

import (
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kirillgashkov/assignment-timetrack/api/timetrackapi/v1"
	"github.com/kirillgashkov/assignment-timetrack/internal/api/request"
	"github.com/kirillgashkov/assignment-timetrack/internal/api/response"
	"github.com/kirillgashkov/assignment-timetrack/internal/config"
)

func NewServer(cfg *config.ServerConfig, db *pgxpool.Pool) (*http.Server, error) {
	h, err := newHandler(db)
	if err != nil {
		return nil, errors.Join(errors.New("failed to create handler"), err)
	}

	return &http.Server{
		Addr:              net.JoinHostPort(cfg.Host, cfg.Port),
		Handler:           h,
		ReadHeaderTimeout: time.Duration(cfg.ReadHeaderTimeout) * time.Second,
	}, nil
}

func newHandler(db *pgxpool.Pool) (http.Handler, error) {
	si := &serverInterface{db: db}
	mux := http.NewServeMux()
	return timetrackapi.HandlerFromMux(si, mux), nil
}

type serverInterface struct {
	db *pgxpool.Pool
}

func (si *serverInterface) GetHealth(w http.ResponseWriter, _ *http.Request) {
	response.MustWriteJSON(w, timetrackapi.Health{Status: "ok"}, http.StatusOK)
}

func (si *serverInterface) PostUsers(w http.ResponseWriter, r *http.Request) {
	var userCreate *timetrackapi.UserCreate
	if err := request.ReadJSON(r, &userCreate); err != nil {
		response.MustWriteError(w, "invalid request", http.StatusUnprocessableEntity)
		return
	}
	if userCreate.PassportNumber == "" {
		response.MustWriteError(w, "missing passport number", http.StatusUnprocessableEntity)
		return
	}

	type userDB struct {
		ID             int
		PassportNumber string `db:"passport_number"`
		Surname        string
		Name           string
		Patronymic     *string
		Address        string
	}

	rows, err := si.db.Query(
		r.Context(),
		`
			INSERT INTO users (passport_number, surname, name, patronymic, address)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, passport_number, surname, name, patronymic, address
		`,
		userCreate.PassportNumber,
		"some surname",
		"some name",
		"some patronymic",
		"some address",
	)
	if err != nil {
		slog.Error("failed to query insert user", "error", err)
		response.MustWriteInternalServerError(w)
		return
	}

	uDB, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[userDB])
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			response.MustWriteError(w, "user with this passport number already exists", http.StatusBadRequest)
			return
		}
		slog.Error("failed to collect rows from insert user", "error", err)
		response.MustWriteInternalServerError(w)
		return
	}

	u := timetrackapi.User{
		Id:             uDB.ID,
		PassportNumber: uDB.PassportNumber,
		Surname:        uDB.Surname,
		Name:           uDB.Name,
		Patronymic:     uDB.Patronymic,
		Address:        uDB.Address,
	}
	response.MustWriteJSON(w, u, http.StatusOK)
}

func (si *serverInterface) GetUsersCurrent(http.ResponseWriter, *http.Request) {
	panic("implement me")
}
