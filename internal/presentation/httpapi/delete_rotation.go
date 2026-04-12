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

type DeleteRotationErrorResponse struct {
	Links  DeleteRotationResponseLinks `json:"links"`
	Errors []ErrorObject               `json:"errors,omitempty"`
}

type DeleteRotationResponseLinks struct {
	Self string `json:"self"`
}

type DeleteRotation = func(ctx context.Context, input application.DeleteRotationInput) error

type DeleteRotationHandler struct {
	hostname       string
	deleteRotation DeleteRotation
}

func NewDeleteRotationHandler(hostname string, deleteRotation DeleteRotation) *DeleteRotationHandler {
	return &DeleteRotationHandler{hostname: hostname, deleteRotation: deleteRotation}
}

func (h *DeleteRotationHandler) Handle(w http.ResponseWriter, r *http.Request) {
	rotationID := r.PathValue("rotationID")
	selfURL := h.hostname + "/api/rotations/" + rotationID

	errorResponse := func(status int, detail string) {
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(DeleteRotationErrorResponse{
			Links: DeleteRotationResponseLinks{Self: selfURL},
			Errors: []ErrorObject{{
				Status: fmt.Sprintf("%d", status),
				Title:  http.StatusText(status),
				Detail: detail,
			}},
		})
	}

	err := h.deleteRotation(r.Context(), application.DeleteRotationInput{
		RotationID: rotationID,
	})
	if errors.Is(err, domain.ErrRotationNotFound) {
		errorResponse(http.StatusNotFound, "Rotation not found")
		return
	}
	if err != nil {
		errorResponse(http.StatusInternalServerError, "An unexpected error occurred")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
