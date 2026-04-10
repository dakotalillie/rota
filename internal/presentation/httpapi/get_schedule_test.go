package httpapi_test

import (
	"context"
	"errors"
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

func TestGetScheduleHandler(t *testing.T) {
	const hostname = "http://localhost:8080"
	const rotationID = "rot_01JQGF0000000000000000000"

	loc, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)

	alice := &domain.Member{
		ID:         "mem_01JQGF0000000000000000000",
		RotationID: rotationID,
		Order:      1,
		Color:      "violet",
		User: domain.User{
			ID:    "usr_01JQGF0000000000000000000",
			Name:  "Alice Smith",
			Email: "alice@example.com",
		},
	}
	bob := &domain.Member{
		ID:         "mem_02JQGF0000000000000000000",
		RotationID: rotationID,
		Order:      2,
		Color:      "sky",
		User: domain.User{
			ID:    "usr_02JQGF0000000000000000000",
			Name:  "Bob Jones",
			Email: "bob@example.com",
		},
	}

	threeBlocks := []domain.ScheduleBlock{
		{
			Start:  time.Date(2026, 3, 30, 9, 0, 0, 0, loc),
			End:    time.Date(2026, 4, 6, 9, 0, 0, 0, loc),
			Member: alice,
		},
		{
			Start:  time.Date(2026, 4, 6, 9, 0, 0, 0, loc),
			End:    time.Date(2026, 4, 13, 9, 0, 0, 0, loc),
			Member: bob,
		},
		{
			Start:  time.Date(2026, 4, 13, 9, 0, 0, 0, loc),
			End:    time.Date(2026, 4, 20, 9, 0, 0, 0, loc),
			Member: alice,
		},
	}

	blocksWithOverride := []domain.ScheduleBlock{
		{
			Start:  time.Date(2026, 3, 30, 9, 0, 0, 0, loc),
			End:    time.Date(2026, 4, 2, 9, 0, 0, 0, loc),
			Member: alice,
		},
		{
			Start:      time.Date(2026, 4, 2, 9, 0, 0, 0, loc),
			End:        time.Date(2026, 4, 4, 9, 0, 0, 0, loc),
			Member:     bob,
			IsOverride: true,
		},
		{
			Start:  time.Date(2026, 4, 4, 9, 0, 0, 0, loc),
			End:    time.Date(2026, 4, 6, 9, 0, 0, 0, loc),
			Member: alice,
		},
	}

	tests := []struct {
		name           string
		url            string
		getter         httpapi.GetSchedule
		wantStatusCode int
	}{
		{
			name: "success",
			url:  "/api/rotations/" + rotationID + "/schedule",
			getter: func(_ context.Context, _ string, _ time.Time, _ int) ([]domain.ScheduleBlock, error) {
				return threeBlocks, nil
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "success - weeks param",
			url:  "/api/rotations/" + rotationID + "/schedule?weeks=3",
			getter: func(_ context.Context, _ string, _ time.Time, numWeeks int) ([]domain.ScheduleBlock, error) {
				require.Equal(t, 3, numWeeks)
				return threeBlocks, nil
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "success - empty schedule",
			url:  "/api/rotations/" + rotationID + "/schedule",
			getter: func(_ context.Context, _ string, _ time.Time, _ int) ([]domain.ScheduleBlock, error) {
				return []domain.ScheduleBlock{}, nil
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "invalid weeks param - not a number",
			url:            "/api/rotations/" + rotationID + "/schedule?weeks=abc",
			getter:         nil,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "invalid weeks param - below range",
			url:            "/api/rotations/" + rotationID + "/schedule?weeks=0",
			getter:         nil,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "invalid weeks param - above range",
			url:            "/api/rotations/" + rotationID + "/schedule?weeks=11",
			getter:         nil,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "not found",
			url:  "/api/rotations/" + rotationID + "/schedule",
			getter: func(_ context.Context, _ string, _ time.Time, _ int) ([]domain.ScheduleBlock, error) {
				return nil, domain.ErrRotationNotFound
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "unprocessable entity",
			url:  "/api/rotations/" + rotationID + "/schedule",
			getter: func(_ context.Context, _ string, _ time.Time, _ int) ([]domain.ScheduleBlock, error) {
				return nil, errors.New("rotation has no weekly cadence")
			},
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "success - with override block",
			url:  "/api/rotations/" + rotationID + "/schedule",
			getter: func(_ context.Context, _ string, _ time.Time, _ int) ([]domain.ScheduleBlock, error) {
				return blocksWithOverride, nil
			},
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := httpapi.NewGetScheduleHandler(hostname, tt.getter, clock.New())

			r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tt.url, nil)
			r.SetPathValue("rotationID", rotationID)
			w := httptest.NewRecorder()

			handler.Handle(w, r)

			require.Equal(t, tt.wantStatusCode, w.Code)
			cupaloy.SnapshotT(t, w.Body.String())
		})
	}
}
