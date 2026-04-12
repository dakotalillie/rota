package httpapi_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dakotalillie/rota/internal/application"
	"github.com/dakotalillie/rota/internal/clock"
	"github.com/dakotalillie/rota/internal/domain"
	"github.com/dakotalillie/rota/internal/presentation/httpapi"
	"github.com/stretchr/testify/require"
)

func TestCreateMemberHandler(t *testing.T) {
	const hostname = "http://localhost:8080"
	const rotationID = "rot_01JQGF0000000000000000000"

	successMember := &domain.Member{
		ID:         "mem_01JQGF0000000000000000001",
		RotationID: rotationID,
		Color:      "violet",
		User: domain.User{
			ID:    "usr_01JQGF0000000000000000000",
			Name:  "Alice Smith",
			Email: "alice@example.com",
		},
		Position: 1,
	}

	tests := []struct {
		name           string
		body           string
		creator        httpapi.CreateMember
		wantStatusCode int
	}{
		{
			name: "success - new user",
			body: `{"data":{"type":"members","attributes":{"name":"Alice Smith","email":"alice@example.com"}}}`,
			creator: func(_ context.Context, _ application.CreateMemberInput) (*domain.Member, error) {
				return successMember, nil
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name: "success - existing user",
			body: `{"data":{"type":"members","relationships":{"user":{"data":{"type":"users","id":"usr_01JQGF0000000000000000000"}}}}}`,
			creator: func(_ context.Context, _ application.CreateMemberInput) (*domain.Member, error) {
				return successMember, nil
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name: "rotation not found",
			body: `{"data":{"type":"members","attributes":{"name":"Alice","email":"alice@example.com"}}}`,
			creator: func(_ context.Context, _ application.CreateMemberInput) (*domain.Member, error) {
				return nil, domain.ErrRotationNotFound
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "user not found",
			body: `{"data":{"type":"members","relationships":{"user":{"data":{"type":"users","id":"usr_99999999999999999999999999"}}}}}`,
			creator: func(_ context.Context, _ application.CreateMemberInput) (*domain.Member, error) {
				return nil, domain.ErrUserNotFound
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "duplicate member",
			body: `{"data":{"type":"members","attributes":{"name":"Alice","email":"alice@example.com"}}}`,
			creator: func(_ context.Context, _ application.CreateMemberInput) (*domain.Member, error) {
				return nil, domain.ErrMemberAlreadyExists
			},
			wantStatusCode: http.StatusConflict,
		},
		{
			name: "rotation full",
			body: `{"data":{"type":"members","attributes":{"name":"Alice","email":"alice@example.com"}}}`,
			creator: func(_ context.Context, _ application.CreateMemberInput) (*domain.Member, error) {
				return nil, domain.ErrRotationMembershipFull
			},
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "missing user fields",
			body: `{"data":{"type":"members","attributes":{"name":"Alice"}}}`,
			creator: func(_ context.Context, _ application.CreateMemberInput) (*domain.Member, error) {
				return nil, domain.ErrMissingUserFields
			},
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name:           "malformed json",
			body:           `{not valid json`,
			creator:        nil,
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := httpapi.NewCreateMemberHandler(hostname, tt.creator, clock.New(), slog.New(slog.NewTextHandler(io.Discard, nil)))

			r := httptest.NewRequestWithContext(t.Context(), http.MethodPost,
				"/api/rotations/"+rotationID+"/members",
				strings.NewReader(tt.body),
			)
			r.SetPathValue("rotationID", rotationID)
			w := httptest.NewRecorder()

			handler.Handle(w, r)

			require.Equal(t, tt.wantStatusCode, w.Code)
			snapshotJSON(t, w.Body.String())
		})
	}
}
