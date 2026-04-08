package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dakotalillie/rota/internal/application"
	"github.com/dakotalillie/rota/internal/domain"
)

type CreateOverrideRequest struct {
	Data CreateOverrideRequestData `json:"data"`
}

type CreateOverrideRequestData struct {
	Attributes    CreateOverrideRequestAttributes    `json:"attributes"`
	Relationships CreateOverrideRequestRelationships `json:"relationships"`
}

type CreateOverrideRequestAttributes struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type CreateOverrideRequestRelationships struct {
	Member *CreateOverrideRequestMemberRelationship `json:"member,omitempty"`
}

type CreateOverrideRequestMemberRelationship struct {
	Data CreateOverrideRequestMemberRelationshipData `json:"data"`
}

type CreateOverrideRequestMemberRelationshipData struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type CreateOverrideResponse struct {
	Links    CreateOverrideResponseLinks `json:"links"`
	Data     *OverrideResource           `json:"data,omitempty"`
	Included []any                       `json:"included,omitempty"`
	Errors   []ErrorObject               `json:"errors,omitempty"`
}

type CreateOverrideResponseLinks struct {
	Self string `json:"self"`
}

type OverrideResource struct {
	Type          string                `json:"type"`
	ID            string                `json:"id"`
	Attributes    OverrideAttributes    `json:"attributes"`
	Relationships OverrideRelationships `json:"relationships"`
}

type OverrideAttributes struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type OverrideRelationships struct {
	Member OverrideMemberRelationship `json:"member"`
}

type OverrideMemberRelationship struct {
	Data OverrideMemberRelationshipData `json:"data"`
}

type OverrideMemberRelationshipData struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type CreateOverride = func(ctx context.Context, input application.CreateOverrideInput) (*domain.Override, error)

type CreateOverrideHandler struct {
	hostname       string
	createOverride CreateOverride
}

func NewCreateOverrideHandler(hostname string, createOverride CreateOverride) *CreateOverrideHandler {
	return &CreateOverrideHandler{hostname: hostname, createOverride: createOverride}
}

func (h *CreateOverrideHandler) Handle(w http.ResponseWriter, r *http.Request) {
	rotationID := r.PathValue("rotationID")
	selfBase := h.hostname + "/api/rotations/" + rotationID + "/overrides"

	errorResponse := func(status int, detail string) {
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(CreateOverrideResponse{
			Links: CreateOverrideResponseLinks{Self: selfBase},
			Errors: []ErrorObject{{
				Status: fmt.Sprintf("%d", status),
				Title:  http.StatusText(status),
				Detail: detail,
			}},
		})
	}

	var req CreateOverrideRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Data.Relationships.Member == nil || req.Data.Relationships.Member.Data.ID == "" {
		errorResponse(http.StatusBadRequest, "member relationship is required")
		return
	}

	start, err := time.Parse(time.RFC3339, req.Data.Attributes.Start)
	if err != nil {
		errorResponse(http.StatusBadRequest, "start must be a valid RFC3339 datetime")
		return
	}

	end, err := time.Parse(time.RFC3339, req.Data.Attributes.End)
	if err != nil {
		errorResponse(http.StatusBadRequest, "end must be a valid RFC3339 datetime")
		return
	}

	if !end.After(start) {
		errorResponse(http.StatusBadRequest, "end must be after start")
		return
	}

	input := application.CreateOverrideInput{
		RotationID: rotationID,
		MemberID:   req.Data.Relationships.Member.Data.ID,
		Start:      start,
		End:        end,
	}

	override, err := h.createOverride(r.Context(), input)
	if errors.Is(err, domain.ErrRotationNotFound) {
		errorResponse(http.StatusNotFound, "Rotation not found")
		return
	}
	if errors.Is(err, domain.ErrMemberNotFound) {
		errorResponse(http.StatusNotFound, "Member not found in rotation")
		return
	}
	if errors.Is(err, domain.ErrOverrideSameMember) {
		errorResponse(http.StatusUnprocessableEntity, "Override member is already the scheduled on-call during this window")
		return
	}
	if errors.Is(err, domain.ErrOverrideConflict) {
		errorResponse(http.StatusConflict, "An override already exists during this window")
		return
	}
	if err != nil {
		errorResponse(http.StatusInternalServerError, "An unexpected error occurred")
		return
	}

	selfURL := selfBase + "/" + override.ID
	w.Header().Set("Location", selfURL)
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(CreateOverrideResponse{
		Links: CreateOverrideResponseLinks{Self: selfURL},
		Data: &OverrideResource{
			Type: "overrides",
			ID:   override.ID,
			Attributes: OverrideAttributes{
				Start: override.Start.Format(time.RFC3339),
				End:   override.End.Format(time.RFC3339),
			},
			Relationships: OverrideRelationships{
				Member: OverrideMemberRelationship{
					Data: OverrideMemberRelationshipData{Type: "members", ID: override.Member.ID},
				},
			},
		},
		Included: []any{
			Member{
				Type:       "members",
				ID:         override.Member.ID,
				Attributes: MemberAttributes{Order: override.Member.Order, Color: override.Member.Color},
				Relationships: MemberRelationships{
					User: MemberUserRelationship{
						Data: MemberUserRelationshipData{Type: "users", ID: override.Member.User.ID},
					},
				},
			},
			IncludedUser{
				Type: "users",
				ID:   override.Member.User.ID,
				Attributes: IncludedUserAttributes{
					Name:  override.Member.User.Name,
					Email: override.Member.User.Email,
				},
			},
		},
	})
}
