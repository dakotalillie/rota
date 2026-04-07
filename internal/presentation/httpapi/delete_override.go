package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/dakotalillie/rota/internal/application"
	"github.com/dakotalillie/rota/internal/domain"
)

type DeleteOverrideErrorResponse struct {
	Links  DeleteOverrideResponseLinks `json:"links"`
	Errors []ErrorObject               `json:"errors,omitempty"`
}

type DeleteOverrideResponseLinks struct {
	Self string `json:"self"`
}

type DeleteOverride = func(ctx context.Context, input application.DeleteOverrideInput) error

type DeleteOverrideHandler struct {
	hostname       string
	deleteOverride DeleteOverride
}

func NewDeleteOverrideHandler(hostname string, deleteOverride DeleteOverride) *DeleteOverrideHandler {
	return &DeleteOverrideHandler{hostname: hostname, deleteOverride: deleteOverride}
}

func (h *DeleteOverrideHandler) Handle(w http.ResponseWriter, r *http.Request) {
	rotationID := r.PathValue("rotationID")
	overrideID := r.PathValue("overrideID")
	selfURL := h.hostname + "/api/rotations/" + rotationID + "/overrides/" + overrideID

	errorResponse := func(status int, detail string) {
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(DeleteOverrideErrorResponse{
			Links: DeleteOverrideResponseLinks{Self: selfURL},
			Errors: []ErrorObject{{
				Status: fmt.Sprintf("%d", status),
				Title:  http.StatusText(status),
				Detail: detail,
			}},
		})
	}

	err := h.deleteOverride(r.Context(), application.DeleteOverrideInput{
		RotationID: rotationID,
		OverrideID: overrideID,
	})
	if errors.Is(err, domain.ErrRotationNotFound) {
		errorResponse(http.StatusNotFound, "Rotation not found")
		return
	}
	if errors.Is(err, domain.ErrOverrideNotFound) {
		errorResponse(http.StatusNotFound, "Override not found")
		return
	}
	if err != nil {
		errorResponse(http.StatusInternalServerError, "An unexpected error occurred")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
