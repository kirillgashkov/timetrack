package tracking

import "net/http"

type Handler struct {
	Service *Service
}

// PostTasksIdStart handles "POST /tasks/{id}/start".
//
//nolint:revive
func (h *Handler) PostTasksIdStart(w http.ResponseWriter, r *http.Request, id int) {
	panic("implement me")
}

// PostTasksIdStop handles "POST /tasks/{id}/stop".
//
//nolint:revive
func (h *Handler) PostTasksIdStop(w http.ResponseWriter, r *http.Request, id int) {
	panic("implement me")
}
