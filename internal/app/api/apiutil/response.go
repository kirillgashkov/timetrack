package apiutil

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/kirillgashkov/timetrack/api/timetrackapi/v1"
)

type ValidationError []string

func (e ValidationError) Error() string {
	sb := strings.Builder{}
	for i, m := range e {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(m)
	}
	return sb.String()
}

func MustWriteJSON(w http.ResponseWriter, v any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		panic(errors.Join(errors.New("failed to write JSON response"), err))
	}
}

func MustWriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func MustWriteError(w http.ResponseWriter, m string, code int) {
	e := timetrackapi.ErrorResponse{Message: m}
	MustWriteJSON(w, e, code)
}

func MustWriteForbidden(w http.ResponseWriter) {
	MustWriteError(w, "forbidden", http.StatusForbidden)
}

func MustWriteInternalServerError(w http.ResponseWriter, m string, e error) {
	e = errors.Join(errors.New(m), e)
	slog.Error("internal server error", "error", e)
	MustWriteError(w, "internal server error", http.StatusInternalServerError)
}

func MustWriteUnprocessableEntity(w http.ResponseWriter, ve ValidationError) {
	MustWriteError(w, ve.Error(), http.StatusUnprocessableEntity)
}

func MustWriteUnauthorized(w http.ResponseWriter, m string) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	MustWriteError(w, m, http.StatusUnauthorized)
}
