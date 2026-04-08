package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/dakotalillie/rota/internal/domain"
	"github.com/dakotalillie/rota/internal/presentation/httpapi"
	"github.com/stretchr/testify/require"
)

func TestListRotationsHandler(t *testing.T) {
	const hostname = "http://localhost:8080"

	rot1 := &domain.Rotation{
		ID:   "rot_01JQGF0000000000000000000",
		Name: "Platform On-Call",
		Cadence: domain.RotationCadence{
			Weekly: &domain.RotationCadenceWeekly{
				Day:      "Monday",
				Time:     "09:00",
				TimeZone: "America/New_York",
			},
		},
	}

	rot2 := &domain.Rotation{
		ID:   "rot_01JQGF1111111111111111111",
		Name: "Database On-Call",
		Cadence: domain.RotationCadence{
			Weekly: &domain.RotationCadenceWeekly{
				Day:      "Tuesday",
				Time:     "10:00",
				TimeZone: "America/Chicago",
			},
		},
		ScheduledMember: &domain.Member{
			ID:         "mem_01JQGF0000000000000000000",
			RotationID: "rot_01JQGF1111111111111111111",
			Order:      1,
			User: domain.User{
				ID:    "usr_01JQGF0000000000000000000",
				Name:  "Alice Smith",
				Email: "alice@example.com",
			},
		},
	}

	tests := []struct {
		name           string
		lister         httpapi.ListRotations
		wantStatusCode int
	}{
		{
			name: "success - empty",
			lister: func(_ context.Context) ([]*domain.Rotation, error) {
				return []*domain.Rotation{}, nil
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "success - no current members",
			lister: func(_ context.Context) ([]*domain.Rotation, error) {
				return []*domain.Rotation{rot1}, nil
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "success - with current member",
			lister: func(_ context.Context) ([]*domain.Rotation, error) {
				return []*domain.Rotation{rot1, rot2}, nil
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "success - deduplicated included",
			lister: func(_ context.Context) ([]*domain.Rotation, error) {
				rot1WithSameMember := *rot1
				rot1WithSameMember.ScheduledMember = rot2.ScheduledMember
				return []*domain.Rotation{&rot1WithSameMember, rot2}, nil
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "success - active override uses effective on-call member",
			lister: func(_ context.Context) ([]*domain.Rotation, error) {
				rot := *rot2
				rot.Overrides = []domain.Override{
					{
						ID:         "ovr_01JQGF0000000000000000001",
						RotationID: rot.ID,
						Start:      time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
						End:        time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC),
						Member: domain.Member{
							ID:         "mem_01JQGF9999999999999999999",
							RotationID: rot.ID,
							Order:      2,
							User: domain.User{
								ID:    "usr_01JQGF9999999999999999999",
								Name:  "Bob Jones",
								Email: "bob@example.com",
							},
						},
					},
				}
				return []*domain.Rotation{&rot}, nil
			},
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := httpapi.NewListRotationsHandler(hostname, tt.lister)

			r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/rotations", nil)
			w := httptest.NewRecorder()

			handler.Handle(w, r)

			require.Equal(t, tt.wantStatusCode, w.Code)
			cupaloy.SnapshotT(t, w.Body.String())
		})
	}
}
