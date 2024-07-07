package request

import (
	"encoding/json"
	"errors"
	"net/http"
)

func ReadJSON(r *http.Request, v any) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return errors.Join(errors.New("failed to decode JSON request"), err)
	}
	return nil
}
