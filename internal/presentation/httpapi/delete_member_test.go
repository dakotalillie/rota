package httpapi_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/dakotalillie/rota/internal/application"
	"github.com/dakotalillie/rota/internal/clock"
	"github.com/dakotalillie/rota/internal/domain"
	"github.com/dakotalillie/rota/internal/presentation/httpapi"
	"github.com/stretchr/testify/require"
)

func TestDeleteMemberHandler(t *testing.T) {
	const hostname = "http://localhost:8080"
	const rotationID = "rot_01JQGF0000000000000000000"
	const memberID = "mem_01JQGF0000000000000000001"

	tests := []struct {
		name           string
		deleter        httpapi.DeleteMember
		wantStatusCode int
	}{
		{
			name: "success",
			deleter: func(_ context.Context, _ application.DeleteMemberInput) error {
				return nil
			},
			wantStatusCode: http.StatusNoContent,
		},
		{
			name: "rotation not found",
			deleter: func(_ context.Context, _ application.DeleteMemberInput) error {
				return domain.ErrRotationNotFound
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "member not found",
			deleter: func(_ context.Context, _ application.DeleteMemberInput) error {
				return domain.ErrMemberNotFound
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "unexpected error",
			deleter: func(_ context.Context, _ application.DeleteMemberInput) error {
				return errors.New("something went wrong")
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := httpapi.NewDeleteMemberHandler(hostname, tt.deleter, clock.New())

			r := httptest.NewRequestWithContext(t.Context(), http.MethodDelete,
				"/api/rotations/"+rotationID+"/members/"+memberID,
				nil,
			)
			r.SetPathValue("rotationID", rotationID)
			r.SetPathValue("memberID", memberID)
			w := httptest.NewRecorder()

			handler.Handle(w, r)

			require.Equal(t, tt.wantStatusCode, w.Code)
			cupaloy.SnapshotT(t, w.Body.String())
		})
	}
}
