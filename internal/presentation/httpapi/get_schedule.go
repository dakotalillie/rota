package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dakotalillie/rota/internal/domain"
)

type GetScheduleResponse struct {
	Links    GetScheduleResponseLinks `json:"links"`
	Data     []ScheduleBlock          `json:"data"`
	Included []any                    `json:"included,omitempty"`
	Errors   []ErrorObject            `json:"errors,omitempty"`
}

type GetScheduleResponseLinks struct {
	Self string `json:"self"`
}

type GetSchedule = func(ctx context.Context, rotationID string, now time.Time, numWeeks int) ([]domain.ScheduleBlock, error)

type GetScheduleHandler struct {
	hostname    string
	getSchedule GetSchedule
}

func NewGetScheduleHandler(hostname string, getSchedule GetSchedule) *GetScheduleHandler {
	return &GetScheduleHandler{hostname: hostname, getSchedule: getSchedule}
}

func (h *GetScheduleHandler) Handle(w http.ResponseWriter, r *http.Request) {
	rotationID := r.PathValue("rotationID")

	numWeeks := 7
	if weeksParam := r.URL.Query().Get("weeks"); weeksParam != "" {
		n, err := strconv.Atoi(weeksParam)
		if err != nil || n < 1 || n > 10 {
			response := GetScheduleResponse{
				Links:  GetScheduleResponseLinks{Self: h.hostname + r.URL.Path},
				Data:   []ScheduleBlock{},
				Errors: []ErrorObject{{Status: "400", Title: "Bad Request", Detail: fmt.Sprintf("weeks must be an integer between 1 and 10, got %q", weeksParam)}},
			}
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(response)
			return
		}
		numWeeks = n
	}

	blocks, err := h.getSchedule(r.Context(), rotationID, time.Now(), numWeeks)
	if errors.Is(err, domain.ErrRotationNotFound) {
		response := GetScheduleResponse{
			Links:  GetScheduleResponseLinks{Self: h.hostname + r.URL.Path},
			Data:   []ScheduleBlock{},
			Errors: []ErrorObject{{Status: "404", Title: "Not Found", Detail: "Rotation not found"}},
		}
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(response)
		return
	}
	if err != nil {
		response := GetScheduleResponse{
			Links:  GetScheduleResponseLinks{Self: h.hostname + r.URL.Path},
			Data:   []ScheduleBlock{},
			Errors: []ErrorObject{{Status: "422", Title: "Unprocessable Entity", Detail: err.Error()}},
		}
		w.WriteHeader(http.StatusUnprocessableEntity)
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	data := make([]ScheduleBlock, len(blocks))
	for i, b := range blocks {
		data[i] = ScheduleBlock{
			Type: "scheduleBlocks",
			ID:   b.Start.Format(time.RFC3339),
			Attributes: ScheduleBlockAttributes{
				Start:      b.Start.Format(time.RFC3339),
				End:        b.End.Format(time.RFC3339),
				IsOverride: b.IsOverride,
			},
			Relationships: ScheduleBlockRelationships{
				Member: ScheduleBlockMemberRelationship{
					Data: RelationshipData{Type: "members", ID: b.Member.ID},
				},
			},
		}
	}

	var included []any
	seenMembers := map[string]bool{}
	seenUsers := map[string]bool{}
	for _, b := range blocks {
		m := b.Member
		if !seenMembers[m.ID] {
			seenMembers[m.ID] = true
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
		}
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

	response := GetScheduleResponse{
		Links:    GetScheduleResponseLinks{Self: h.hostname + r.URL.Path},
		Data:     data,
		Included: included,
	}
	_ = json.NewEncoder(w).Encode(response)
}
