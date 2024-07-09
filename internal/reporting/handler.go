package reporting

import (
	"net/http"
	"time"

	"github.com/kirillgashkov/timetrack/api/timetrackapi/v1"
	"github.com/kirillgashkov/timetrack/internal/app/api/apiutil"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

// PostUsersIdReport handles "POST /users/{id}/report".
//
//nolint:revive
func (h *Handler) PostUsersIdReport(
	w http.ResponseWriter, r *http.Request, id int, params timetrackapi.PostUsersIdReportParams,
) {
	var from time.Time
	if params.From != nil {
		from = *params.From
	} else {
		apiutil.MustWriteError(w, "from is required", http.StatusUnprocessableEntity)
	}
	var to time.Time
	if params.To != nil {
		to = *params.To
	} else {
		apiutil.MustWriteError(w, "to is required", http.StatusUnprocessableEntity)

	}

	report, err := h.Service.Report(r.Context(), id, from, to)
	if err != nil {
		apiutil.MustWriteInternalServerError(w)
		return
	}

	apiutil.MustWriteJSON(w, report, http.StatusOK)
}
