package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

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
	clock       domain.Clock
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

	membersData := make([]RelationshipData, 0, len(rotation.Members))
	included := make([]any, 0)
	seenUsers := map[string]bool{}
	for _, m := range rotation.Members {
		membersData = append(membersData, RelationshipData{Type: "members", ID: m.ID})
		included = append(included, Member{
			Type:       "members",
			ID:         m.ID,
			Attributes: MemberAttributes{Order: m.Order, Color: m.Color},
			Relationships: MemberRelationships{
				User: MemberUserRelationship{
					Data: MemberUserRelationshipData{Type: "users", ID: m.User.ID},
				},
			},
		})
		if !seenUsers[m.User.ID] {
			seenUsers[m.User.ID] = true
			included = append(included, IncludedUser{
				Type: "users",
				ID:   m.User.ID,
				Attributes: IncludedUserAttributes{
					Name:  m.User.Name,
					Email: m.User.Email,
				},
			})
		}
	}
	response.Data.Relationships.Members = &MembersRelationship{Data: membersData}

	overridesData := make([]RelationshipData, 0, len(rotation.Overrides))
	for _, o := range rotation.Overrides {
		overridesData = append(overridesData, RelationshipData{Type: "overrides", ID: o.ID})
		included = append(included, OverrideResource{
			Type: "overrides",
			ID:   o.ID,
			Attributes: OverrideAttributes{
				Start: o.Start.Format(time.RFC3339),
				End:   o.End.Format(time.RFC3339),
			},
			Relationships: OverrideRelationships{
				Member: OverrideMemberRelationship{
					Data: OverrideMemberRelationshipData{Type: "members", ID: o.Member.ID},
				},
			},
		})
	}
	response.Data.Relationships.Overrides = &OverridesRelationship{Data: overridesData}
	response.Included = included

	if currentMember := rotation.EffectiveOnCall(h.clock.Now()); currentMember != nil {
		response.Data.Relationships.CurrentMember.Data = &RelationshipData{
			Type: "members",
			ID:   currentMember.ID,
		}
	}
	if rotation.ScheduledMember != nil {
		response.Data.Relationships.ScheduledMember.Data = &RelationshipData{
			Type: "members",
			ID:   rotation.ScheduledMember.ID,
		}
	}

	_ = json.NewEncoder(w).Encode(response)
}

func NewGetRotationHandler(hostname string, getRotation GetRotation, clock domain.Clock) *GetRotationHandler {
	return &GetRotationHandler{hostname: hostname, getRotation: getRotation, clock: clock}
}
