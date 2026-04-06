package httpapi_test

import (
	"context"
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

func TestCreateRotationHandler(t *testing.T) {
	const hostname = "http://localhost:8080"

	successRotation := &domain.Rotation{
		ID:   "rot_01JQGF0000000000000000000",
		Name: "Platform On-Call",
		Cadence: domain.RotationCadence{
			Weekly: &domain.RotationCadenceWeekly{
				Day:      "monday",
				Time:     "09:00",
				TimeZone: "America/Los_Angeles",
			},
		},
	}

	tests := []struct {
		name           string
		body           string
		creator        httpapi.CreateRotation
		wantStatusCode int
	}{
		{
			name: "success",
			body: `{"data":{"type":"rotations","attributes":{"name":"Platform On-Call"}}}`,
			creator: func(_ context.Context, _ application.CreateRotationInput) (*domain.Rotation, error) {
				return successRotation, nil
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name: "missing name",
			body: `{"data":{"type":"rotations","attributes":{"name":""}}}`,
			creator: func(_ context.Context, _ application.CreateRotationInput) (*domain.Rotation, error) {
				return nil, domain.ErrInvalidRotationName
			},
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "too many rotations",
			body: `{"data":{"type":"rotations","attributes":{"name":"New Rotation"}}}`,
			creator: func(_ context.Context, _ application.CreateRotationInput) (*domain.Rotation, error) {
				return nil, domain.ErrTooManyRotations
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
			handler := httpapi.NewCreateRotationHandler(hostname, tt.creator)

			r := httptest.NewRequestWithContext(t.Context(), http.MethodPost,
				"/api/rotations",
				strings.NewReader(tt.body),
			)
			w := httptest.NewRecorder()

			handler.Handle(w, r)

			require.Equal(t, tt.wantStatusCode, w.Code)
			cupaloy.SnapshotT(t, w.Body.String())
		})
	}
}
