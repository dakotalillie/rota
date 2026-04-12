package httpapi_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dakotalillie/rota/internal/application"
	"github.com/dakotalillie/rota/internal/domain"
	"github.com/dakotalillie/rota/internal/presentation/httpapi"
	"github.com/stretchr/testify/require"
)

func TestDeleteOverrideHandler(t *testing.T) {
	const hostname = "http://localhost:8080"
	const rotationID = "rot_01JQGF0000000000000000000"
	const overrideID = "ovr_01JQGF0000000000000000001"

	tests := []struct {
		name           string
		deleter        httpapi.DeleteOverride
		wantStatusCode int
	}{
		{
			name: "success",
			deleter: func(_ context.Context, _ application.DeleteOverrideInput) error {
				return nil
			},
			wantStatusCode: http.StatusNoContent,
		},
		{
			name: "rotation not found",
			deleter: func(_ context.Context, _ application.DeleteOverrideInput) error {
				return domain.ErrRotationNotFound
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "override not found",
			deleter: func(_ context.Context, _ application.DeleteOverrideInput) error {
				return domain.ErrOverrideNotFound
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "unexpected error",
			deleter: func(_ context.Context, _ application.DeleteOverrideInput) error {
				return errors.New("something went wrong")
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := httpapi.NewDeleteOverrideHandler(hostname, tt.deleter)

			r := httptest.NewRequestWithContext(t.Context(), http.MethodDelete,
				"/api/rotations/"+rotationID+"/overrides/"+overrideID,
				nil,
			)
			r.SetPathValue("rotationID", rotationID)
			r.SetPathValue("overrideID", overrideID)
			w := httptest.NewRecorder()

			handler.Handle(w, r)

			require.Equal(t, tt.wantStatusCode, w.Code)
			snapshotJSON(t, w.Body.String())
		})
	}
}
