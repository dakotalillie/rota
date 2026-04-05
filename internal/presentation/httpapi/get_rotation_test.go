package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/dakotalillie/rota/internal/domain"
	"github.com/dakotalillie/rota/internal/presentation/httpapi"
	"github.com/stretchr/testify/require"
)

func TestGetRotationHandler(t *testing.T) {
	const hostname = "http://localhost:8080"
	const rotationID = "rot_01JQGF0000000000000000000"

	tests := []struct {
		name           string
		getter         httpapi.RotationGetter
		wantStatusCode int
	}{
		{
			name: "success",
			getter: func(id string) (*domain.Rotation, error) {
				return &domain.Rotation{
					ID:   rotationID,
					Name: "Platform On-Call",
					Cadence: domain.RotationCadence{
						Weekly: &domain.RotationCadenceWeekly{
							Day:      "Monday",
							Time:     "09:00",
							TimeZone: "America/New_York",
						},
					},
				}, nil
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "not found",
			getter: func(id string) (*domain.Rotation, error) {
				return nil, domain.ErrRotationNotFound
			},
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := httpapi.NewGetRotationHandler(hostname, tt.getter)

			r := httptest.NewRequest(http.MethodGet, "/api/rotations/"+rotationID, nil)
			r.SetPathValue("rotationID", rotationID)
			w := httptest.NewRecorder()

			handler.Handle(w, r)

			require.Equal(t, tt.wantStatusCode, w.Code)
			cupaloy.SnapshotT(t, w.Body.String())
		})
	}
}
