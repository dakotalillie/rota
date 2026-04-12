package httpapi_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dakotalillie/rota/internal/application"
	"github.com/dakotalillie/rota/internal/domain"
	"github.com/dakotalillie/rota/internal/presentation/httpapi"
	"github.com/stretchr/testify/require"
)

func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func TestCreateOverrideHandler(t *testing.T) {
	const hostname = "http://localhost:8080"
	const rotationID = "rot_01JQGF0000000000000000000"

	successOverride := &domain.Override{
		ID:         "ovr_01JQGF0000000000000000001",
		RotationID: rotationID,
		Member: domain.Member{
			ID:    "mem_01JQGF0000000000000000001",
			Color: "violet",
			User: domain.User{
				ID:    "usr_01JQGF0000000000000000000",
				Name:  "Alice Smith",
				Email: "alice@example.com",
			},
			Position: 1,
		},
		Start: mustParseTime("2026-04-07T09:00:00Z"),
		End:   mustParseTime("2026-04-14T09:00:00Z"),
	}

	validBody := `{"data":{"type":"overrides","attributes":{"start":"2026-04-07T09:00:00Z","end":"2026-04-14T09:00:00Z"},"relationships":{"member":{"data":{"type":"members","id":"mem_01JQGF0000000000000000001"}}}}}`

	tests := []struct {
		name           string
		body           string
		creator        httpapi.CreateOverride
		wantStatusCode int
	}{
		{
			name: "success",
			body: validBody,
			creator: func(_ context.Context, _ application.CreateOverrideInput) (*domain.Override, error) {
				return successOverride, nil
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name: "rotation not found",
			body: validBody,
			creator: func(_ context.Context, _ application.CreateOverrideInput) (*domain.Override, error) {
				return nil, domain.ErrRotationNotFound
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "member not found",
			body: validBody,
			creator: func(_ context.Context, _ application.CreateOverrideInput) (*domain.Override, error) {
				return nil, domain.ErrMemberNotFound
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "override same member",
			body: validBody,
			creator: func(_ context.Context, _ application.CreateOverrideInput) (*domain.Override, error) {
				return nil, domain.ErrOverrideSameMember
			},
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "override conflict",
			body: validBody,
			creator: func(_ context.Context, _ application.CreateOverrideInput) (*domain.Override, error) {
				return nil, domain.ErrOverrideConflict
			},
			wantStatusCode: http.StatusConflict,
		},
		{
			name: "unexpected error",
			body: validBody,
			creator: func(_ context.Context, _ application.CreateOverrideInput) (*domain.Override, error) {
				return nil, errors.New("db exploded")
			},
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:           "malformed json",
			body:           `{not valid json`,
			creator:        nil,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "missing member relationship",
			body:           `{"data":{"type":"overrides","attributes":{"start":"2026-04-07T09:00:00Z","end":"2026-04-14T09:00:00Z"}}}`,
			creator:        nil,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "invalid start time",
			body:           `{"data":{"type":"overrides","attributes":{"start":"not-a-time","end":"2026-04-14T09:00:00Z"},"relationships":{"member":{"data":{"type":"members","id":"mem_01JQGF0000000000000000001"}}}}}`,
			creator:        nil,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "invalid end time",
			body:           `{"data":{"type":"overrides","attributes":{"start":"2026-04-07T09:00:00Z","end":"not-a-time"},"relationships":{"member":{"data":{"type":"members","id":"mem_01JQGF0000000000000000001"}}}}}`,
			creator:        nil,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "end before start",
			body:           `{"data":{"type":"overrides","attributes":{"start":"2026-04-14T09:00:00Z","end":"2026-04-07T09:00:00Z"},"relationships":{"member":{"data":{"type":"members","id":"mem_01JQGF0000000000000000001"}}}}}`,
			creator:        nil,
			wantStatusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := httpapi.NewCreateOverrideHandler(hostname, tt.creator, slog.New(slog.NewTextHandler(io.Discard, nil)))

			r := httptest.NewRequestWithContext(t.Context(), http.MethodPost,
				"/api/rotations/"+rotationID+"/overrides",
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
