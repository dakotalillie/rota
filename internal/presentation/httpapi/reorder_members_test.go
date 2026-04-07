package httpapi_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/dakotalillie/rota/internal/application"
	"github.com/dakotalillie/rota/internal/domain"
	"github.com/dakotalillie/rota/internal/presentation/httpapi"
	"github.com/stretchr/testify/require"
)

func TestReorderMembersHandler(t *testing.T) {
	const hostname = "http://localhost:8080"
	const rotationID = "rot_01JQGF0000000000000000000"

	successRotation := &domain.Rotation{
		ID:   rotationID,
		Name: "Platform On-Call",
		Members: []domain.Member{
			{
				ID:    "mem_01JQGF0000000000000000003",
				Order: 1,
				User:  domain.User{ID: "usr_01JQGF0000000000000000003", Name: "Charlie", Email: "charlie@example.com"},
			},
			{
				ID:    "mem_01JQGF0000000000000000001",
				Order: 2,
				User:  domain.User{ID: "usr_01JQGF0000000000000000001", Name: "Alice", Email: "alice@example.com"},
			},
			{
				ID:    "mem_01JQGF0000000000000000002",
				Order: 3,
				User:  domain.User{ID: "usr_01JQGF0000000000000000002", Name: "Bob", Email: "bob@example.com"},
			},
		},
	}

	tests := []struct {
		name           string
		body           string
		reorderer      httpapi.ReorderMembers
		wantStatusCode int
	}{
		{
			name: "success",
			body: `{"data":[{"type":"members","id":"mem_01JQGF0000000000000000003"},{"type":"members","id":"mem_01JQGF0000000000000000001"},{"type":"members","id":"mem_01JQGF0000000000000000002"}]}`,
			reorderer: func(_ context.Context, _ application.ReorderMembersInput) (*domain.Rotation, error) {
				return successRotation, nil
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "rotation not found",
			body: `{"data":[{"type":"members","id":"mem_01JQGF0000000000000000001"}]}`,
			reorderer: func(_ context.Context, _ application.ReorderMembersInput) (*domain.Rotation, error) {
				return nil, domain.ErrRotationNotFound
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "member mismatch",
			body: `{"data":[{"type":"members","id":"mem_01JQGF0000000000000000001"},{"type":"members","id":"mem_unknown"}]}`,
			reorderer: func(_ context.Context, _ application.ReorderMembersInput) (*domain.Rotation, error) {
				return nil, domain.ErrMemberMismatch
			},
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name:           "malformed json",
			body:           `{not valid json`,
			reorderer:      nil,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "unexpected error",
			body: `{"data":[{"type":"members","id":"mem_01JQGF0000000000000000001"}]}`,
			reorderer: func(_ context.Context, _ application.ReorderMembersInput) (*domain.Rotation, error) {
				return nil, errors.New("something went wrong")
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := httpapi.NewReorderMembersHandler(hostname, tt.reorderer)

			r := httptest.NewRequestWithContext(t.Context(), http.MethodPut,
				"/api/rotations/"+rotationID+"/members",
				strings.NewReader(tt.body),
			)
			r.SetPathValue("rotationID", rotationID)
			w := httptest.NewRecorder()

			handler.Handle(w, r)

			require.Equal(t, tt.wantStatusCode, w.Code)
			cupaloy.SnapshotT(t, w.Body.String())
		})
	}
}
