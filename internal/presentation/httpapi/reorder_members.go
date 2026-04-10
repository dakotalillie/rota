package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/dakotalillie/rota/internal/application"
	"github.com/dakotalillie/rota/internal/domain"
)

type ReorderMembersRequest struct {
	Data []ReorderMembersRequestItem `json:"data"`
}

type ReorderMembersRequestItem struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type ReorderMembersResponse struct {
	Links    ReorderMembersResponseLinks `json:"links"`
	Data     []Member                    `json:"data,omitempty"`
	Included []IncludedUser              `json:"included,omitempty"`
	Errors   []ErrorObject               `json:"errors,omitempty"`
}

type ReorderMembersResponseLinks struct {
	Self string `json:"self"`
}

type ReorderMembers = func(ctx context.Context, input application.ReorderMembersInput) (*domain.Rotation, error)

type ReorderMembersHandler struct {
	hostname       string
	reorderMembers ReorderMembers
}

func NewReorderMembersHandler(hostname string, reorderMembers ReorderMembers) *ReorderMembersHandler {
	return &ReorderMembersHandler{hostname: hostname, reorderMembers: reorderMembers}
}

func (h *ReorderMembersHandler) Handle(w http.ResponseWriter, r *http.Request) {
	rotationID := r.PathValue("rotationID")
	selfURL := h.hostname + "/api/rotations/" + rotationID + "/members"

	errorResponse := func(status int, detail string) {
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(ReorderMembersResponse{
			Links: ReorderMembersResponseLinks{Self: selfURL},
			Errors: []ErrorObject{{
				Status: fmt.Sprintf("%d", status),
				Title:  http.StatusText(status),
				Detail: detail,
			}},
		})
	}

	var req ReorderMembersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(http.StatusBadRequest, "Invalid request body")
		return
	}

	memberIDs := make([]string, len(req.Data))
	for i, item := range req.Data {
		memberIDs[i] = item.ID
	}

	rotation, err := h.reorderMembers(r.Context(), application.ReorderMembersInput{
		RotationID: rotationID,
		MemberIDs:  memberIDs,
	})
	if errors.Is(err, domain.ErrRotationNotFound) {
		errorResponse(http.StatusNotFound, "Rotation not found")
		return
	}
	if errors.Is(err, domain.ErrMemberMismatch) {
		errorResponse(http.StatusUnprocessableEntity, "The provided members do not match the rotation's current members")
		return
	}
	if err != nil {
		errorResponse(http.StatusInternalServerError, "An unexpected error occurred")
		return
	}

	members := make([]Member, 0, len(rotation.Members))
	seenUsers := map[string]bool{}
	included := make([]IncludedUser, 0, len(rotation.Members))

	sorted := make([]domain.Member, len(rotation.Members))
	copy(sorted, rotation.Members)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Position < sorted[j].Position
	})

	for _, m := range sorted {
		members = append(members, Member{
			Type:       "members",
			ID:         m.ID,
			Attributes: MemberAttributes{Position: m.Position, Color: m.Color},
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

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ReorderMembersResponse{
		Links:    ReorderMembersResponseLinks{Self: selfURL},
		Data:     members,
		Included: included,
	})
}
