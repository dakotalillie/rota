package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dakotalillie/rota/internal/application"
	"github.com/dakotalillie/rota/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteOverrideUseCase_Execute(t *testing.T) {
	const rotationID = "rot_01JQGF0000000000000000000"
	const overrideID = "ovr_01JQGF0000000000000000001"

	tests := []struct {
		name             string
		input            application.DeleteOverrideInput
		rotationRepo     fakeRotationRepo
		overrideRepo     fakeOverrideRepo
		wantErr          error
		wantDeleteCall   bool
		wantDeleteRotID  string
		wantDeleteOverID string
	}{
		{
			name:             "success",
			input:            application.DeleteOverrideInput{RotationID: rotationID, OverrideID: overrideID},
			rotationRepo:     fakeRotationRepo{rotation: &domain.Rotation{ID: rotationID, Name: "Platform On-Call"}},
			wantDeleteCall:   true,
			wantDeleteRotID:  rotationID,
			wantDeleteOverID: overrideID,
		},
		{
			name:         "rotation not found",
			input:        application.DeleteOverrideInput{RotationID: rotationID, OverrideID: overrideID},
			rotationRepo: fakeRotationRepo{err: domain.ErrRotationNotFound},
			wantErr:      domain.ErrRotationNotFound,
		},
		{
			name:             "override not found",
			input:            application.DeleteOverrideInput{RotationID: rotationID, OverrideID: overrideID},
			rotationRepo:     fakeRotationRepo{rotation: &domain.Rotation{ID: rotationID, Name: "Platform On-Call"}},
			overrideRepo:     fakeOverrideRepo{deleteErr: domain.ErrOverrideNotFound},
			wantErr:          domain.ErrOverrideNotFound,
			wantDeleteCall:   true,
			wantDeleteRotID:  rotationID,
			wantDeleteOverID: overrideID,
		},
		{
			name:             "delete error propagates",
			input:            application.DeleteOverrideInput{RotationID: rotationID, OverrideID: overrideID},
			rotationRepo:     fakeRotationRepo{rotation: &domain.Rotation{ID: rotationID, Name: "Platform On-Call"}},
			overrideRepo:     fakeOverrideRepo{deleteErr: errors.New("db error")},
			wantErr:          errors.New("db error"),
			wantDeleteCall:   true,
			wantDeleteRotID:  rotationID,
			wantDeleteOverID: overrideID,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := application.NewDeleteOverrideUseCase(
				&fakeTransactor{},
				&tc.rotationRepo,
				&tc.overrideRepo,
			)

			err := uc.Execute(context.Background(), tc.input)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, tc.wantErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			if tc.wantDeleteCall {
				require.Len(t, tc.overrideRepo.deleteCalls, 1)
				assert.Equal(t, tc.wantDeleteRotID, tc.overrideRepo.deleteCalls[0].rotationID)
				assert.Equal(t, tc.wantDeleteOverID, tc.overrideRepo.deleteCalls[0].overrideID)
			} else {
				assert.Empty(t, tc.overrideRepo.deleteCalls)
			}
		})
	}
}
