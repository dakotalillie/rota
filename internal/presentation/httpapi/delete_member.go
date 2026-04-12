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

type DeleteMemberErrorResponse struct {
	Links  DeleteMemberResponseLinks `json:"links"`
	Errors []ErrorObject             `json:"errors,omitempty"`
}

type DeleteMemberResponseLinks struct {
	Self string `json:"self"`
}

type DeleteMember = func(ctx context.Context, input application.DeleteMemberInput) error

type DeleteMemberHandler struct {
	hostname     string
	deleteMember DeleteMember
	clock        domain.Clock
	logger       *slog.Logger
}

func NewDeleteMemberHandler(hostname string, deleteMember DeleteMember, clock domain.Clock, logger *slog.Logger) *DeleteMemberHandler {
	return &DeleteMemberHandler{hostname: hostname, deleteMember: deleteMember, clock: clock, logger: logger}
}

func (h *DeleteMemberHandler) Handle(w http.ResponseWriter, r *http.Request) {
	rotationID := r.PathValue("rotationID")
	memberID := r.PathValue("memberID")
	selfURL := h.hostname + "/api/rotations/" + rotationID + "/members/" + memberID

	errorResponse := func(status int, detail string) {
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(DeleteMemberErrorResponse{
			Links: DeleteMemberResponseLinks{Self: selfURL},
			Errors: []ErrorObject{{
				Status: fmt.Sprintf("%d", status),
				Title:  http.StatusText(status),
				Detail: detail,
			}},
		})
	}

	err := h.deleteMember(r.Context(), application.DeleteMemberInput{
		RotationID: rotationID,
		MemberID:   memberID,
		Now:        h.clock.Now(),
	})
	if errors.Is(err, domain.ErrRotationNotFound) {
		errorResponse(http.StatusNotFound, "Rotation not found")
		return
	}
	if errors.Is(err, domain.ErrMemberNotFound) {
		errorResponse(http.StatusNotFound, "Member not found")
		return
	}
	if err != nil {
		h.logger.Error("failed to remove member", "rotation_id", rotationID, "member_id", memberID, "error", err)
		errorResponse(http.StatusInternalServerError, "An unexpected error occurred")
		return
	}

	h.logger.Info("member removed", "rotation_id", rotationID, "member_id", memberID)

	w.WriteHeader(http.StatusNoContent)
}
