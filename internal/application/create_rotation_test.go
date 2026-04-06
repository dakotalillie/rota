package application_test

import (
	"context"
	"testing"

	"github.com/dakotalillie/rota/internal/application"
	"github.com/dakotalillie/rota/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateRotationUseCase_Execute(t *testing.T) {
	tests := []struct {
		name         string
		input        application.CreateRotationInput
		rotationRepo fakeRotationRepo
		wantErr      error
		wantRotation func(t *testing.T, got *domain.Rotation)
	}{
		{
			name:  "success",
			input: application.CreateRotationInput{Name: "Platform On-Call"},
			wantRotation: func(t *testing.T, got *domain.Rotation) {
				t.Helper()
				assert.Equal(t, "Platform On-Call", got.Name)
				require.NotNil(t, got.Cadence.Weekly)
				assert.Equal(t, "monday", got.Cadence.Weekly.Day)
				assert.Equal(t, "09:00", got.Cadence.Weekly.Time)
				assert.Equal(t, "America/Los_Angeles", got.Cadence.Weekly.TimeZone)
			},
		},
		{
			name:    "empty name",
			input:   application.CreateRotationInput{Name: ""},
			wantErr: domain.ErrInvalidRotationName,
		},
		{
			name:         "rotation limit reached",
			input:        application.CreateRotationInput{Name: "New Rotation"},
			rotationRepo: fakeRotationRepo{count: 20},
			wantErr:      domain.ErrTooManyRotations,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := application.NewCreateRotationUseCase(&fakeTransactor{}, &tt.rotationRepo)
			got, err := uc.Execute(context.Background(), tt.input)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			tt.wantRotation(t, got)
		})
	}
}
