package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/dakotalillie/rota/internal/clock"
	"github.com/dakotalillie/rota/internal/domain"
	"github.com/dakotalillie/rota/internal/presentation/httpapi"
	"github.com/stretchr/testify/require"
)

func TestGetRotationHandler(t *testing.T) {
	const hostname = "http://localhost:8080"
	const rotationID = "rot_01JQGF0000000000000000000"

	tests := []struct {
		name           string
		getter         httpapi.GetRotation
		wantStatusCode int
	}{
		{
			name: "success",
			getter: func(_ context.Context, id string) (*domain.Rotation, error) {
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
			name: "success - with current member",
			getter: func(_ context.Context, id string) (*domain.Rotation, error) {
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
					ScheduledMember: &domain.Member{
						ID:         "mem_01JQGF0000000000000000000",
						RotationID: rotationID,
						Position:   1,
						Color:      "violet",
						User: domain.User{
							ID:    "usr_01JQGF0000000000000000000",
							Name:  "Alice Smith",
							Email: "alice@example.com",
						},
					},
					Members: []domain.Member{
						{
							ID:         "mem_01JQGF0000000000000000000",
							RotationID: rotationID,
							Position:   1,
							Color:      "violet",
							User: domain.User{
								ID:    "usr_01JQGF0000000000000000000",
								Name:  "Alice Smith",
								Email: "alice@example.com",
							},
						},
						{
							ID:         "mem_01JQGF000000000000000000B",
							RotationID: rotationID,
							Position:   2,
							Color:      "sky",
							User: domain.User{
								ID:    "usr_01JQGF000000000000000000B",
								Name:  "Bob Jones",
								Email: "bob@example.com",
							},
						},
					},
				}, nil
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "success - with members",
			getter: func(_ context.Context, id string) (*domain.Rotation, error) {
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
					Members: []domain.Member{
						{
							ID:         "mem_01JQGF0000000000000000000",
							RotationID: rotationID,
							Position:   1,
							Color:      "violet",
							User: domain.User{
								ID:    "usr_01JQGF0000000000000000000",
								Name:  "Alice Smith",
								Email: "alice@example.com",
							},
						},
						{
							ID:         "mem_01JQGF000000000000000000B",
							RotationID: rotationID,
							Position:   2,
							Color:      "sky",
							User: domain.User{
								ID:    "usr_01JQGF000000000000000000B",
								Name:  "Bob Jones",
								Email: "bob@example.com",
							},
						},
					},
				}, nil
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "success - with overrides",
			getter: func(_ context.Context, id string) (*domain.Rotation, error) {
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
					Members: []domain.Member{
						{
							ID:         "mem_01JQGF0000000000000000000",
							RotationID: rotationID,
							Position:   1,
							Color:      "violet",
							User: domain.User{
								ID:    "usr_01JQGF0000000000000000000",
								Name:  "Alice Smith",
								Email: "alice@example.com",
							},
						},
						{
							ID:         "mem_01JQGF000000000000000000B",
							RotationID: rotationID,
							Position:   2,
							Color:      "sky",
							User: domain.User{
								ID:    "usr_01JQGF000000000000000000B",
								Name:  "Bob Jones",
								Email: "bob@example.com",
							},
						},
					},
					Overrides: []domain.Override{
						{
							ID:         "ovr_01JQGF0000000000000000001",
							RotationID: rotationID,
							Start:      time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
							End:        time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC),
							Member: domain.Member{
								ID:         "mem_01JQGF000000000000000000B",
								RotationID: rotationID,
								Position:   2,
								Color:      "sky",
								User: domain.User{
									ID:    "usr_01JQGF000000000000000000B",
									Name:  "Bob Jones",
									Email: "bob@example.com",
								},
							},
						},
					},
				}, nil
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "not found",
			getter: func(_ context.Context, id string) (*domain.Rotation, error) {
				return nil, domain.ErrRotationNotFound
			},
			wantStatusCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := httpapi.NewGetRotationHandler(hostname, tt.getter, clock.New())

			r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/rotations/"+rotationID, nil)
			r.SetPathValue("rotationID", rotationID)
			w := httptest.NewRecorder()

			handler.Handle(w, r)

			require.Equal(t, tt.wantStatusCode, w.Code)
			cupaloy.SnapshotT(t, w.Body.String())
		})
	}
}
