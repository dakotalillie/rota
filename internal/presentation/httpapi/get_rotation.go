package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dakotalillie/rota/internal/domain"
)

type GetRotationResponse struct {
	Links    GetRotationResponseLinks `json:"links"`
	Data     *Rotation                `json:"data,omitempty"`
	Included []any                    `json:"included,omitempty"`
	Errors   []ErrorObject            `json:"errors,omitempty"`
}

type GetRotationResponseLinks struct {
	Self string `json:"self"`
}

type GetRotation = func(ctx context.Context, id string) (*domain.Rotation, error)

type GetRotationHandler struct {
	hostname    string
	getRotation GetRotation
}

func (h *GetRotationHandler) Handle(w http.ResponseWriter, r *http.Request) {
	rotationID := r.PathValue("rotationID")

	rotation, err := h.getRotation(r.Context(), rotationID)
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
		_ = json.NewEncoder(w).Encode(response)
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
						Day:      rotation.Cadence.Weekly.Day,
						Time:     rotation.Cadence.Weekly.Time,
						TimeZone: rotation.Cadence.Weekly.TimeZone,
					},
				},
			},
		},
	}

	if rotation.CurrentMember != nil {
		cm := rotation.CurrentMember
		response.Data.Relationships.CurrentMember.Data = &RelationshipData{
			Type: "members",
			ID:   cm.ID,
		}
		response.Included = []any{
			Member{
				Type:       "members",
				ID:         cm.ID,
				Attributes: MemberAttributes{Order: cm.Order},
				Relationships: MemberRelationships{
					User: MemberUserRelationship{
						Data: MemberUserRelationshipData{Type: "users", ID: cm.User.ID},
					},
				},
			},
			IncludedUser{
				Type: "users",
				ID:   cm.User.ID,
				Attributes: IncludedUserAttributes{
					Name:  cm.User.Name,
					Email: cm.User.Email,
				},
			},
		}
	}

	_ = json.NewEncoder(w).Encode(response)
}

func NewGetRotationHandler(hostname string, getRotation GetRotation) *GetRotationHandler {
	return &GetRotationHandler{hostname: hostname, getRotation: getRotation}
}
