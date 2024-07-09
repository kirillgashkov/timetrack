package apiutil

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/kirillgashkov/timetrack/api/timetrackapi/v1"
)

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
	e := timetrackapi.Error{Message: m}
	MustWriteJSON(w, e, code)
}

func MustWriteInternalServerError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	_, err := w.Write([]byte("internal server error"))
	if err != nil {
		panic(errors.Join(errors.New("failed to write response"), err))
	}
}
