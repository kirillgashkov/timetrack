package response

import (
	"encoding/json"
	"errors"
	"net/http"
)

func MustWriteJSON(w http.ResponseWriter, v any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		panic(errors.Join(errors.New("failed to write JSON response"), err))
	}
}

func MustWriteInternalServerError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	_, err := w.Write([]byte("internal server error"))
	if err != nil {
		panic(errors.Join(errors.New("failed to write response"), err))
	}
}
