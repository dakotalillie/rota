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

func TestListRotationsUseCase_Execute(t *testing.T) {
	rotation1 := &domain.Rotation{ID: "rot_1", Name: "Platform"}
	rotation2 := &domain.Rotation{ID: "rot_2", Name: "Database"}
	override := domain.Override{
		ID:         "ovr_1",
		RotationID: rotation2.ID,
		Member:     domain.Member{ID: "mem_2"},
	}

	t.Run("loads overrides for each rotation", func(t *testing.T) {
		overrideRepo := fakeOverrideRepo{
			listByRotationID: map[string][]domain.Override{
				rotation1.ID: []domain.Override{},
				rotation2.ID: []domain.Override{override},
			},
		}
		rotationRepo := fakeRotationRepo{rotations: []*domain.Rotation{rotation1, rotation2}}

		uc := application.NewListRotationsUseCase(&rotationRepo, &overrideRepo)

		got, err := uc.Execute(context.Background())

		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Empty(t, got[0].Overrides)
		assert.Equal(t, []domain.Override{override}, got[1].Overrides)
		assert.Empty(t, overrideRepo.listByRotationCalls)
		require.Len(t, overrideRepo.listByRotationIDsCalls, 1)
		assert.Equal(t, []string{rotation1.ID, rotation2.ID}, overrideRepo.listByRotationIDsCalls[0].rotationIDs)
	})

	t.Run("list error propagates", func(t *testing.T) {
		rotationRepo := fakeRotationRepo{err: errors.New("db error")}

		uc := application.NewListRotationsUseCase(&rotationRepo, &fakeOverrideRepo{})

		got, err := uc.Execute(context.Background())

		require.Error(t, err)
		assert.Equal(t, "db error", err.Error())
		assert.Nil(t, got)
	})

	t.Run("override error propagates", func(t *testing.T) {
		rotationRepo := fakeRotationRepo{rotations: []*domain.Rotation{rotation1}}
		overrideRepo := fakeOverrideRepo{listByRotationErr: errors.New("override error")}

		uc := application.NewListRotationsUseCase(&rotationRepo, &overrideRepo)

		got, err := uc.Execute(context.Background())

		require.Error(t, err)
		assert.Equal(t, "override error", err.Error())
		assert.Nil(t, got)
	})
}
