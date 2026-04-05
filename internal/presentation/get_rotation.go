package presentation

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dakotalillie/rota/internal/domain"
)

type GetRotationResponse struct {
	Links  GetRotationResponseLinks `json:"links"`
	Data   *Rotation                `json:"data,omitempty"`
	Errors []ErrorObject            `json:"errors,omitempty"`
}

type GetRotationResponseLinks struct {
	Self string `json:"self"`
}

type RotationGetter = func(id string) (*domain.Rotation, error)

type GetRotationHandler struct {
	hostname    string
	getRotation RotationGetter
}

func (h *GetRotationHandler) Handle(w http.ResponseWriter, r *http.Request) {
	rotationID := r.PathValue("rotationID")

	rotation, err := h.getRotation(rotationID)
	if errors.Is(err, domain.ErrRotationNotFound) {
		response := GetRotationResponse{
			Links: GetRotationResponseLinks{
				Self: h.hostname + r.URL.Path,
			},
			Errors: []ErrorObject{
				{
					Status: "404",
					Title:  "Not Found",
					Detail: "Rotation not found",
				},
			},
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := GetRotationResponse{
		Links: GetRotationResponseLinks{
			Self: h.hostname + r.URL.Path,
		},
		Data: &Rotation{
			Type: "rotations",
			ID:   rotation.ID,
			Attributes: RotationAttributes{
				Name: rotation.Name,
				Cadence: RotationCadence{
					Weekly: &RotationCadenceWeekly{
						RotationDay:      rotation.Cadence.Weekly.RotationDay,
						RotationTime:     rotation.Cadence.Weekly.RotationTime,
						RotationTimeZone: rotation.Cadence.Weekly.RotationTimeZone,
					},
				},
			},
		},
	}
	json.NewEncoder(w).Encode(response)
}

func NewGetRotationHandler(hostname string, getRotation RotationGetter) *GetRotationHandler {
	return &GetRotationHandler{hostname: hostname, getRotation: getRotation}
}
