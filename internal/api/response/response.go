package response

import (
	"encoding/json"
	"net/http"
)

func MustWriteJSON(w http.ResponseWriter, v interface{}) {
	if err := WriteJSON(w, v); err != nil {
		panic(err)
	}
}

func WriteJSON(w http.ResponseWriter, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}
