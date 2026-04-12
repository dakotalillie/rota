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

type CreateMemberRequest struct {
	Data CreateMemberRequestData `json:"data"`
}

type CreateMemberRequestData struct {
	Attributes    CreateMemberRequestAttributes    `json:"attributes"`
	Relationships CreateMemberRequestRelationships `json:"relationships"`
}

type CreateMemberRequestAttributes struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateMemberRequestRelationships struct {
	User *CreateMemberRequestUserRelationship `json:"user,omitempty"`
}

type CreateMemberRequestUserRelationship struct {
	Data CreateMemberRequestUserRelationshipData `json:"data"`
}

type CreateMemberRequestUserRelationshipData struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type CreateMemberResponse struct {
	Links    CreateMemberResponseLinks `json:"links"`
	Data     *Member                   `json:"data,omitempty"`
	Included []IncludedUser            `json:"included,omitempty"`
	Errors   []ErrorObject             `json:"errors,omitempty"`
}

type CreateMemberResponseLinks struct {
	Self string `json:"self"`
}

type CreateMember = func(ctx context.Context, input application.CreateMemberInput) (*domain.Member, error)

type CreateMemberHandler struct {
	hostname     string
	createMember CreateMember
	clock        domain.Clock
	logger       *slog.Logger
}

func NewCreateMemberHandler(hostname string, createMember CreateMember, clock domain.Clock, logger *slog.Logger) *CreateMemberHandler {
	return &CreateMemberHandler{hostname: hostname, createMember: createMember, clock: clock, logger: logger}
}

func (h *CreateMemberHandler) Handle(w http.ResponseWriter, r *http.Request) {
	rotationID := r.PathValue("rotationID")
	selfBase := h.hostname + "/api/rotations/" + rotationID + "/members"

	errorResponse := func(status int, detail string) {
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(CreateMemberResponse{
			Links: CreateMemberResponseLinks{Self: selfBase},
			Errors: []ErrorObject{{
				Status: fmt.Sprintf("%d", status),
				Title:  http.StatusText(status),
				Detail: detail,
			}},
		})
	}

	var req CreateMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(http.StatusBadRequest, "Invalid request body")
		return
	}

	input := application.CreateMemberInput{RotationID: rotationID, Now: h.clock.Now()}
	if req.Data.Relationships.User != nil {
		input.UserID = req.Data.Relationships.User.Data.ID
	} else {
		input.UserName = req.Data.Attributes.Name
		input.UserEmail = req.Data.Attributes.Email
	}

	member, err := h.createMember(r.Context(), input)
	if errors.Is(err, domain.ErrRotationNotFound) {
		errorResponse(http.StatusNotFound, "Rotation not found")
		return
	}
	if errors.Is(err, domain.ErrUserNotFound) {
		errorResponse(http.StatusNotFound, "User not found")
		return
	}
	if errors.Is(err, domain.ErrMemberAlreadyExists) {
		errorResponse(http.StatusConflict, "User is already a member of this rotation")
		return
	}
	if errors.Is(err, domain.ErrRotationMembershipFull) {
		errorResponse(http.StatusUnprocessableEntity, "Rotation has reached the maximum number of members (20)")
		return
	}
	if errors.Is(err, domain.ErrMissingUserFields) {
		errorResponse(http.StatusUnprocessableEntity, "name and email are required when not specifying a user ID")
		return
	}
	if err != nil {
		h.logger.Error("failed to add member", "rotation_id", rotationID, "error", err)
		errorResponse(http.StatusInternalServerError, "An unexpected error occurred")
		return
	}

	h.logger.Info("member added", "rotation_id", rotationID, "member_id", member.ID)

	selfURL := selfBase + "/" + member.ID
	w.Header().Set("Location", selfURL)
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(CreateMemberResponse{
		Links: CreateMemberResponseLinks{Self: selfURL},
		Data: &Member{
			Type:       "members",
			ID:         member.ID,
			Attributes: MemberAttributes{Position: member.Position, Color: member.Color},
			Relationships: MemberRelationships{
				User: MemberUserRelationship{
					Data: MemberUserRelationshipData{Type: "users", ID: member.User.ID},
				},
			},
		},
		Included: []IncludedUser{
			{
				Type: "users",
				ID:   member.User.ID,
				Attributes: IncludedUserAttributes{
					Name:  member.User.Name,
					Email: member.User.Email,
				},
			},
		},
	})
}
