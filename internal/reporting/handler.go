package reporting

import (
	"net/http"

	"github.com/kirillgashkov/timetrack/api/timetrackapi/v1"
	"github.com/kirillgashkov/timetrack/internal/app/api/apiutil"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// PostUsersIdReport handles "POST /users/{id}/report".
//
//nolint:revive
func (h *Handler) PostUsersIdReport(w http.ResponseWriter, r *http.Request, id int) {
	var reportIn *timetrackapi.ReportIn
	if err := apiutil.ReadJSON(r, &reportIn); err != nil {
		apiutil.MustWriteError(w, "invalid request", http.StatusUnprocessableEntity)
		return
	}
	if reportIn.From.IsZero() {
		apiutil.MustWriteError(w, "missing from", http.StatusUnprocessableEntity)
		return
	}
	if reportIn.To.IsZero() {
		apiutil.MustWriteError(w, "missing to", http.StatusUnprocessableEntity)
		return
	}

	report, err := h.service.Report(r.Context(), id, reportIn.From, reportIn.To)
	if err != nil {
		apiutil.MustWriteInternalServerError(w)
		return
	}

	apiutil.MustWriteJSON(w, report, http.StatusOK)
}
