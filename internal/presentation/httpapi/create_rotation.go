package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/dakotalillie/rota/internal/application"
	"github.com/dakotalillie/rota/internal/domain"
)

type CreateRotationRequest struct {
	Data CreateRotationRequestData `json:"data"`
}

type CreateRotationRequestData struct {
	Attributes CreateRotationRequestAttributes `json:"attributes"`
}

type CreateRotationRequestAttributes struct {
	Name string `json:"name"`
}

type CreateRotationResponse struct {
	Links  CreateRotationResponseLinks `json:"links"`
	Data   *Rotation                   `json:"data,omitempty"`
	Errors []ErrorObject               `json:"errors,omitempty"`
}

type CreateRotationResponseLinks struct {
	Self string `json:"self"`
}

type CreateRotation = func(ctx context.Context, input application.CreateRotationInput) (*domain.Rotation, error)

type CreateRotationHandler struct {
	hostname       string
	createRotation CreateRotation
	logger         *slog.Logger
}

func NewCreateRotationHandler(hostname string, createRotation CreateRotation, logger *slog.Logger) *CreateRotationHandler {
	return &CreateRotationHandler{hostname: hostname, createRotation: createRotation, logger: logger}
}

func (h *CreateRotationHandler) Handle(w http.ResponseWriter, r *http.Request) {
	selfBase := h.hostname + "/api/rotations"

	errorResponse := func(status int, detail string) {
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(CreateRotationResponse{
			Links: CreateRotationResponseLinks{Self: selfBase},
			Errors: []ErrorObject{{
				Status: fmt.Sprintf("%d", status),
				Title:  http.StatusText(status),
				Detail: detail,
			}},
		})
	}

	var req CreateRotationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(http.StatusBadRequest, "Invalid request body")
		return
	}

	rotation, err := h.createRotation(r.Context(), application.CreateRotationInput{
		Name: req.Data.Attributes.Name,
	})
	if errors.Is(err, domain.ErrInvalidRotationName) {
		errorResponse(http.StatusUnprocessableEntity, "rotation name is required")
		return
	}
	if errors.Is(err, domain.ErrTooManyRotations) {
		errorResponse(http.StatusUnprocessableEntity, "maximum number of rotations reached (20)")
		return
	}
	if err != nil {
		h.logger.Error("failed to create rotation", "error", err)
		errorResponse(http.StatusInternalServerError, "An unexpected error occurred")
		return
	}

	h.logger.Info("rotation created", "rotation_id", rotation.ID, "name", rotation.Name)

	selfURL := selfBase + "/" + rotation.ID
	w.Header().Set("Location", selfURL)
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(CreateRotationResponse{
		Links: CreateRotationResponseLinks{Self: selfURL},
		Data: &Rotation{
			Type:  "rotations",
			ID:    rotation.ID,
			Links: RotationLinks{Self: selfURL, Schedule: selfURL + "/schedule"},
			Attributes: RotationAttributes{
				Name: rotation.Name,
				Cadence: RotationCadence{
					Weekly: &RotationCadenceWeekly{
						Day:      rotation.Cadence.Weekly.Day,
						Time:     rotation.Cadence.Weekly.Time,
						TimeZone: rotation.Cadence.Weekly.TimeZone,
					},
				},
			},
			Relationships: RotationRelationships{
				CurrentMember: CurrentMemberRelationship{Data: nil},
			},
		},
	})
}
