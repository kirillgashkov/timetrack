package response

import (
	"encoding/json"
	"net/http"
)

func MustWriteJSON(w http.ResponseWriter, v any, code int) {
	if err := WriteJSON(w, v, code); err != nil {
		panic(err)
	}
}

func WriteJSON(w http.ResponseWriter, v any, code int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(v)
}
