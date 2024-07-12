// Package main is the entry point of a pseudo-external service that provides
// information about people.
//
// The main TimeTrack service uses this service to fetch information about users
// on user creation. In a real-world scenario, this service would be implemented
// somewhere else.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/kirillgashkov/timetrack/api/peopleinfoapi/v1"
	"github.com/kirillgashkov/timetrack/internal/app/config"
	"github.com/kirillgashkov/timetrack/internal/app/logging"
)

func main() {
	if err := mainErr(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func mainErr() error {
	cfg, err := config.New()
	if err != nil {
		return errors.Join(errors.New("failed to create config"), err)
	}

	logger := logging.NewLogger(cfg)
	slog.SetDefault(logger)

	srv, err := newServer(&cfg.Server)
	if err != nil {
		return errors.Join(errors.New("failed to create server"), err)
	}

	slog.Info("starting server", "addr", srv.Addr, "mode", cfg.Mode)
	if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return errors.Join(errors.New("failed to listen and serve"), err)
	}

	return nil
}

func newServer(cfg *config.ServerConfig) (*http.Server, error) {
	si := newHandler()
	mux := http.NewServeMux()
	h := peopleinfoapi.HandlerFromMux(si, mux)

	srv := &http.Server{
		Addr:              net.JoinHostPort(cfg.Host, cfg.Port),
		Handler:           h,
		ReadHeaderTimeout: time.Duration(cfg.ReadHeaderTimeout) * time.Second,
	}
	return srv, nil
}

type handler struct {
	names       []string
	patronymics []string
	surnames    []string
	addresses   []string
}

func newHandler() *handler {
	names := []string{
		"Александр",
		"Иван",
		"Дмитрий",
		"Сергей",
		"Михаил",
		"Андрей",
		"Владимир",
		"Николай",
		"Артем",
		"Алексей",
		"Павел",
		"Кирилл",
		"Максим",
		"Виктор",
		"Георгий",
		"Егор",
		"Роман",
		"Игорь",
		"Василий",
		"Тимофей",
	}
	patronymics := []string{
		"Александрович",
		"Иванович",
		"Сергеевич",
		"Владимирович",
		"Дмитриевич",
		"Петрович",
		"Николаевич",
		"Андреевич",
		"Алексеевич",
		"Михайлович",
		"Егорович",
		"Ярославович",
		"Олегович",
		"Евгеньевич",
		"Артемович",
		"Витальевич",
		"Игоревич",
		"Григорьевич",
		"Даниилович",
		"Тимофеевич",
	}
	surnames := []string{
		"Иванов",
		"Петров",
		"Смирнов",
		"Кузнецов",
		"Васильев",
		"Попов",
		"Соколов",
		"Михайлов",
		"Новиков",
		"Федоров",
		"Морозов",
		"Волков",
		"Алексеев",
		"Лебедев",
		"Семёнов",
		"Егоров",
		"Павлов",
		"Козлов",
		"Степанов",
		"Никитин",
	}
	addresses := []string{
		"ул. Цветочная, д. 7",
		"пр. Пионерский, д. 23",
		"пер. Лунный, д. 14",
		"ул. Солнечная, д. 5",
		"пр. Сосновый, д. 31",
		"ул. Звездная, д. 10",
		"пер. Радужный, д. 4",
		"ул. Весенняя, д. 18",
		"пр. Лесной, д. 9",
		"ул. Луговая, д. 12",
		"пер. Речной, д. 27",
		"пр. Озерный, д. 3",
		"ул. Полевая, д. 6",
		"пер. Тихий, д. 11",
		"ул. Ветеранов, д. 22",
		"пр. Школьный, д. 17",
		"ул. Парковая, д. 8",
		"пер. Чернышевского, д. 2",
		"ул. Дружбы, д. 13",
		"пр. Комсомольский, д. 19",
	}

	return &handler{
		names:       names,
		patronymics: patronymics,
		surnames:    surnames,
		addresses:   addresses,
	}
}

// GetInfo handles "GET /info".
func (h *handler) GetInfo(w http.ResponseWriter, _ *http.Request, params peopleinfoapi.GetInfoParams) {
	// Validate input.

	if params.PassportSerie < 0 || params.PassportSerie > 9999 {
		mustWriteJSON(w, map[string]string{"error": "invalid passport serie"}, http.StatusBadRequest)
		return
	}
	if params.PassportNumber < 0 || params.PassportNumber > 999999 {
		mustWriteJSON(w, map[string]string{"error": "invalid passport number"}, http.StatusBadRequest)
		return
	}

	// If passport serie is 404, return "404 Not Found".

	if params.PassportSerie == 404 {
		mustWriteJSON(w, map[string]string{"error": "info not found"}, http.StatusNotFound)
		return
	}

	// Otherwise, generate random person info using the provided passport number
	// and serie as random seed to make the results deterministic.

	src := rand.NewPCG(uint64(params.PassportSerie), uint64(params.PassportNumber))
	//nolint:gosec // This is not a cryptographic use case.
	rnd := rand.New(src)

	name := h.names[rnd.IntN(len(h.names))]
	var patronymic *string
	if rnd.IntN(6) == 0 {
		patronymic = &h.patronymics[rnd.IntN(len(h.patronymics))]
	}
	surname := h.surnames[rnd.IntN(len(h.surnames))]
	address := h.addresses[rnd.IntN(len(h.addresses))]

	// Return the generated person info.

	person := &peopleinfoapi.People{
		Name:       name,
		Patronymic: patronymic,
		Surname:    surname,
		Address:    address,
	}
	mustWriteJSON(w, person, http.StatusOK)
}

func mustWriteJSON(w http.ResponseWriter, v any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		panic(errors.Join(errors.New("failed to write JSON response"), err))
	}
}
