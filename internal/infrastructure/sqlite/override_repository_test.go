package sqlite_test

import (
	"testing"
	"time"

	"github.com/dakotalillie/rota/internal/domain"
	"github.com/dakotalillie/rota/internal/infrastructure/sqlite"
	"github.com/stretchr/testify/require"
)

func TestOverrideRepository_Create(t *testing.T) {
	db := openTestDB(t)
	rotRepo := sqlite.NewRotationRepository(db)
	userRepo := sqlite.NewUserRepository(db)
	memberRepo := sqlite.NewMemberRepository(db)
	overrideRepo := sqlite.NewOverrideRepository(db)

	require.NoError(t, rotRepo.UpsertRotation(t.Context(), rotationA))
	user, err := userRepo.Create(t.Context(), "Alice Smith", "alice@example.com")
	require.NoError(t, err)
	member, err := memberRepo.Create(t.Context(), rotationA.ID, user.ID, 1, domain.MemberColors[0])
	require.NoError(t, err)

	start := time.Date(2026, 4, 7, 9, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 14, 9, 0, 0, 0, time.UTC)

	override, err := overrideRepo.Create(t.Context(), rotationA.ID, member.ID, start, end)
	require.NoError(t, err)
	require.NotEmpty(t, override.ID)
	require.Equal(t, rotationA.ID, override.RotationID)
	require.Equal(t, member.ID, override.Member.ID)
	require.Equal(t, start, override.Start)
	require.Equal(t, end, override.End)
}

func TestOverrideRepository_HasOverlapping(t *testing.T) {
	baseStart := time.Date(2026, 4, 7, 9, 0, 0, 0, time.UTC)
	baseEnd := time.Date(2026, 4, 14, 9, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		queryStart  time.Time
		queryEnd    time.Time
		wantOverlap bool
	}{
		{
			name:        "no overlap - query entirely before existing",
			queryStart:  baseStart.AddDate(0, 0, -14),
			queryEnd:    baseStart.AddDate(0, 0, -7),
			wantOverlap: false,
		},
		{
			name:        "no overlap - query entirely after existing",
			queryStart:  baseEnd.AddDate(0, 0, 7),
			queryEnd:    baseEnd.AddDate(0, 0, 14),
			wantOverlap: false,
		},
		{
			name:        "no overlap - query ends exactly at existing start",
			queryStart:  baseStart.AddDate(0, 0, -7),
			queryEnd:    baseStart,
			wantOverlap: false,
		},
		{
			name:        "no overlap - query starts exactly at existing end",
			queryStart:  baseEnd,
			queryEnd:    baseEnd.AddDate(0, 0, 7),
			wantOverlap: false,
		},
		{
			name:        "overlap - identical window",
			queryStart:  baseStart,
			queryEnd:    baseEnd,
			wantOverlap: true,
		},
		{
			name:        "overlap - query contains existing",
			queryStart:  baseStart.AddDate(0, 0, -1),
			queryEnd:    baseEnd.AddDate(0, 0, 1),
			wantOverlap: true,
		},
		{
			name:        "overlap - existing contains query",
			queryStart:  baseStart.AddDate(0, 0, 1),
			queryEnd:    baseEnd.AddDate(0, 0, -1),
			wantOverlap: true,
		},
		{
			name:        "overlap - partial overlap at start",
			queryStart:  baseStart.AddDate(0, 0, -3),
			queryEnd:    baseStart.AddDate(0, 0, 3),
			wantOverlap: true,
		},
		{
			name:        "overlap - partial overlap at end",
			queryStart:  baseEnd.AddDate(0, 0, -3),
			queryEnd:    baseEnd.AddDate(0, 0, 3),
			wantOverlap: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openTestDB(t)
			rotRepo := sqlite.NewRotationRepository(db)
			userRepo := sqlite.NewUserRepository(db)
			memberRepo := sqlite.NewMemberRepository(db)
			overrideRepo := sqlite.NewOverrideRepository(db)

			require.NoError(t, rotRepo.UpsertRotation(t.Context(), rotationA))
			user, err := userRepo.Create(t.Context(), "Alice Smith", "alice@example.com")
			require.NoError(t, err)
			member, err := memberRepo.Create(t.Context(), rotationA.ID, user.ID, 1, domain.MemberColors[0])
			require.NoError(t, err)

			_, err = overrideRepo.Create(t.Context(), rotationA.ID, member.ID, baseStart, baseEnd)
			require.NoError(t, err)

			got, err := overrideRepo.HasOverlapping(t.Context(), rotationA.ID, tt.queryStart, tt.queryEnd)
			require.NoError(t, err)
			require.Equal(t, tt.wantOverlap, got)
		})
	}
}

func TestOverrideRepository_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db := openTestDB(t)
		rotRepo := sqlite.NewRotationRepository(db)
		userRepo := sqlite.NewUserRepository(db)
		memberRepo := sqlite.NewMemberRepository(db)
		overrideRepo := sqlite.NewOverrideRepository(db)

		require.NoError(t, rotRepo.UpsertRotation(t.Context(), rotationA))
		user, err := userRepo.Create(t.Context(), "Alice Smith", "alice@example.com")
		require.NoError(t, err)
		member, err := memberRepo.Create(t.Context(), rotationA.ID, user.ID, 1, domain.MemberColors[0])
		require.NoError(t, err)

		start := time.Date(2026, 4, 7, 9, 0, 0, 0, time.UTC)
		end := time.Date(2026, 4, 14, 9, 0, 0, 0, time.UTC)

		override, err := overrideRepo.Create(t.Context(), rotationA.ID, member.ID, start, end)
		require.NoError(t, err)

		require.NoError(t, overrideRepo.Delete(t.Context(), rotationA.ID, override.ID))

		overrides, err := overrideRepo.ListByRotationID(t.Context(), rotationA.ID, start.Add(-time.Hour))
		require.NoError(t, err)
		require.Empty(t, overrides)
	})

	t.Run("not found", func(t *testing.T) {
		db := openTestDB(t)
		rotRepo := sqlite.NewRotationRepository(db)
		overrideRepo := sqlite.NewOverrideRepository(db)

		require.NoError(t, rotRepo.UpsertRotation(t.Context(), rotationA))

		err := overrideRepo.Delete(t.Context(), rotationA.ID, "ovr_99999999999999999999999999")
		require.ErrorIs(t, err, domain.ErrOverrideNotFound)
	})

	t.Run("wrong rotation", func(t *testing.T) {
		db := openTestDB(t)
		rotRepo := sqlite.NewRotationRepository(db)
		userRepo := sqlite.NewUserRepository(db)
		memberRepo := sqlite.NewMemberRepository(db)
		overrideRepo := sqlite.NewOverrideRepository(db)

		rotationB := &domain.Rotation{ID: "rot_01JQGF0000000000000000001", Name: "Other Rotation"}
		require.NoError(t, rotRepo.UpsertRotation(t.Context(), rotationA))
		require.NoError(t, rotRepo.UpsertRotation(t.Context(), rotationB))
		user, err := userRepo.Create(t.Context(), "Alice Smith", "alice@example.com")
		require.NoError(t, err)
		member, err := memberRepo.Create(t.Context(), rotationA.ID, user.ID, 1, domain.MemberColors[0])
		require.NoError(t, err)

		start := time.Date(2026, 4, 7, 9, 0, 0, 0, time.UTC)
		end := time.Date(2026, 4, 14, 9, 0, 0, 0, time.UTC)
		override, err := overrideRepo.Create(t.Context(), rotationA.ID, member.ID, start, end)
		require.NoError(t, err)

		err = overrideRepo.Delete(t.Context(), rotationB.ID, override.ID)
		require.ErrorIs(t, err, domain.ErrOverrideNotFound)
	})
}
