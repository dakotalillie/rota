package httpapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/dakotalillie/rota/internal/domain"
)

type ListRotationsResponse struct {
	Links    ListRotationsResponseLinks `json:"links"`
	Data     []*Rotation                `json:"data"`
	Included []any                      `json:"included,omitempty"`
	Errors   []ErrorObject              `json:"errors,omitempty"`
}

type ListRotationsResponseLinks struct {
	Self string `json:"self"`
}

type ListRotations = func(ctx context.Context) ([]*domain.Rotation, error)

type ListRotationsHandler struct {
	hostname      string
	listRotations ListRotations
}

func (h *ListRotationsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	rotations, err := h.listRotations(r.Context())
	if err != nil {
		response := ListRotationsResponse{
			Links:  ListRotationsResponseLinks{Self: h.hostname + r.URL.Path},
			Errors: []ErrorObject{{Status: "500", Title: "Internal Server Error", Detail: err.Error()}},
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	response := ListRotationsResponse{
		Links: ListRotationsResponseLinks{Self: h.hostname + r.URL.Path},
		Data:  make([]*Rotation, 0, len(rotations)),
	}

	seenMembers := map[string]bool{}
	seenUsers := map[string]bool{}

	for _, rot := range rotations {
		resource := &Rotation{
			Type: "rotations",
			ID:   rot.ID,
			Attributes: RotationAttributes{
				Name: rot.Name,
				Cadence: RotationCadence{
					Weekly: &RotationCadenceWeekly{
						Day:      rot.Cadence.Weekly.Day,
						Time:     rot.Cadence.Weekly.Time,
						TimeZone: rot.Cadence.Weekly.TimeZone,
					},
				},
			},
		}

		if rot.CurrentMember != nil {
			cm := rot.CurrentMember
			resource.Relationships.CurrentMember.Data = &RelationshipData{
				Type: "members",
				ID:   cm.ID,
			}
			if !seenMembers[cm.ID] {
				seenMembers[cm.ID] = true
				response.Included = append(response.Included, Member{
					Type:       "members",
					ID:         cm.ID,
					Attributes: MemberAttributes{Order: cm.Order, Color: cm.Color},
					Relationships: MemberRelationships{
						User: MemberUserRelationship{
							Data: MemberUserRelationshipData{Type: "users", ID: cm.User.ID},
						},
					},
				})
			}
			if !seenUsers[cm.User.ID] {
				seenUsers[cm.User.ID] = true
				response.Included = append(response.Included, IncludedUser{
					Type: "users",
					ID:   cm.User.ID,
					Attributes: IncludedUserAttributes{
						Name:  cm.User.Name,
						Email: cm.User.Email,
					},
				})
			}
		}

		response.Data = append(response.Data, resource)
	}

	_ = json.NewEncoder(w).Encode(response)
}

func NewListRotationsHandler(hostname string, listRotations ListRotations) *ListRotationsHandler {
	return &ListRotationsHandler{hostname: hostname, listRotations: listRotations}
}
