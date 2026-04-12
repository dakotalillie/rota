package httpapi_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dakotalillie/rota/internal/application"
	"github.com/dakotalillie/rota/internal/domain"
	"github.com/dakotalillie/rota/internal/presentation/httpapi"
	"github.com/stretchr/testify/require"
)

func TestDeleteRotationHandler(t *testing.T) {
	const hostname = "http://localhost:8080"
	const rotationID = "rot_01JQGF0000000000000000000"

	tests := []struct {
		name           string
		deleter        httpapi.DeleteRotation
		wantStatusCode int
	}{
		{
			name: "success",
			deleter: func(_ context.Context, _ application.DeleteRotationInput) error {
				return nil
			},
			wantStatusCode: http.StatusNoContent,
		},
		{
			name: "rotation not found",
			deleter: func(_ context.Context, _ application.DeleteRotationInput) error {
				return domain.ErrRotationNotFound
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "unexpected error",
			deleter: func(_ context.Context, _ application.DeleteRotationInput) error {
				return errors.New("something went wrong")
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := httpapi.NewDeleteRotationHandler(hostname, tt.deleter, slog.New(slog.NewTextHandler(io.Discard, nil)))

			r := httptest.NewRequestWithContext(t.Context(), http.MethodDelete,
				"/api/rotations/"+rotationID,
				nil,
			)
			r.SetPathValue("rotationID", rotationID)
			w := httptest.NewRecorder()

			handler.Handle(w, r)

			require.Equal(t, tt.wantStatusCode, w.Code)
			snapshotJSON(t, w.Body.String())
		})
	}
}
