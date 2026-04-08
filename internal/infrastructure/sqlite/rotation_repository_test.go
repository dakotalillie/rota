package sqlite_test

import (
	"database/sql"
	"testing"

	"github.com/dakotalillie/rota/internal/domain"
	"github.com/dakotalillie/rota/internal/infrastructure/sqlite"
	"github.com/stretchr/testify/require"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sqlite.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

var rotationA = &domain.Rotation{
	ID:   "rot_01JQGF0000000000000000000",
	Name: "Platform On-Call",
	Cadence: domain.RotationCadence{
		Weekly: &domain.RotationCadenceWeekly{
			Day:      "Monday",
			Time:     "09:00",
			TimeZone: "America/Los_Angeles",
		},
	},
}

func TestRotationRepository_GetRotationByID(t *testing.T) {
	tests := []struct {
		name    string
		seed    *domain.Rotation
		queryID string
		wantRot *domain.Rotation
		wantErr error
	}{
		{
			name:    "not found - empty database",
			seed:    nil,
			queryID: "rot_any",
			wantRot: nil,
			wantErr: domain.ErrRotationNotFound,
		},
		{
			name:    "not found - wrong id",
			seed:    rotationA,
			queryID: "rot_99999999999999999999999999",
			wantRot: nil,
			wantErr: domain.ErrRotationNotFound,
		},
		{
			name:    "success",
			seed:    rotationA,
			queryID: rotationA.ID,
			wantRot: rotationA,
			wantErr: nil,
		},
		{
			name: "success - no weekly cadence",
			seed: &domain.Rotation{
				ID:      "rot_01JQGF0000000000000000001",
				Name:    "No Cadence Rotation",
				Cadence: domain.RotationCadence{Weekly: nil},
			},
			queryID: "rot_01JQGF0000000000000000001",
			wantRot: &domain.Rotation{
				ID:      "rot_01JQGF0000000000000000001",
				Name:    "No Cadence Rotation",
				Cadence: domain.RotationCadence{Weekly: nil},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openTestDB(t)
			repo := sqlite.NewRotationRepository(db)
			if tt.seed != nil {
				require.NoError(t, repo.UpsertRotation(t.Context(), tt.seed))
			}
			got, err := repo.GetByID(t.Context(), tt.queryID)
			require.ErrorIs(t, err, tt.wantErr)
			require.Equal(t, tt.wantRot, got)
		})
	}
}

func TestRotationRepository_UpsertRotation(t *testing.T) {
	tests := []struct {
		name    string
		seed    *domain.Rotation
		upsert  *domain.Rotation
		wantRot *domain.Rotation
	}{
		{
			name:    "insert new rotation",
			seed:    nil,
			upsert:  rotationA,
			wantRot: rotationA,
		},
		{
			name: "update existing rotation",
			seed: &domain.Rotation{
				ID:   rotationA.ID,
				Name: "Original",
				Cadence: domain.RotationCadence{
					Weekly: &domain.RotationCadenceWeekly{
						Day:      "Monday",
						Time:     "09:00",
						TimeZone: "America/Los_Angeles",
					},
				},
			},
			upsert: &domain.Rotation{
				ID:   rotationA.ID,
				Name: "Updated",
				Cadence: domain.RotationCadence{
					Weekly: &domain.RotationCadenceWeekly{
						Day:      "Tuesday",
						Time:     "10:00",
						TimeZone: "America/New_York",
					},
				},
			},
			wantRot: &domain.Rotation{
				ID:   rotationA.ID,
				Name: "Updated",
				Cadence: domain.RotationCadence{
					Weekly: &domain.RotationCadenceWeekly{
						Day:      "Tuesday",
						Time:     "10:00",
						TimeZone: "America/New_York",
					},
				},
			},
		},
		{
			name:   "round trip fidelity",
			seed:   nil,
			upsert: rotationA,
			wantRot: &domain.Rotation{
				ID:   rotationA.ID,
				Name: rotationA.Name,
				Cadence: domain.RotationCadence{
					Weekly: &domain.RotationCadenceWeekly{
						Day:      rotationA.Cadence.Weekly.Day,
						Time:     rotationA.Cadence.Weekly.Time,
						TimeZone: rotationA.Cadence.Weekly.TimeZone,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openTestDB(t)
			repo := sqlite.NewRotationRepository(db)
			if tt.seed != nil {
				require.NoError(t, repo.UpsertRotation(t.Context(), tt.seed))
			}
			require.NoError(t, repo.UpsertRotation(t.Context(), tt.upsert))
			got, err := repo.GetByID(t.Context(), tt.upsert.ID)
			require.NoError(t, err)
			require.Equal(t, tt.wantRot, got)
		})
	}
}

func TestRotationRepository_GetRotationByID_WithScheduledMember(t *testing.T) {
	db := openTestDB(t)
	rotRepo := sqlite.NewRotationRepository(db)
	userRepo := sqlite.NewUserRepository(db)
	memberRepo := sqlite.NewMemberRepository(db)

	require.NoError(t, rotRepo.UpsertRotation(t.Context(), rotationA))
	user, err := userRepo.Create(t.Context(), "Alice Smith", "alice@example.com")
	require.NoError(t, err)
	member, err := memberRepo.Create(t.Context(), rotationA.ID, user.ID, 1)
	require.NoError(t, err)

	_, err = db.ExecContext(t.Context(),
		`UPDATE members SET is_current = 1, became_current_at = '2026-04-01T09:00:00Z' WHERE id = ?`,
		member.ID,
	)
	require.NoError(t, err)

	got, err := rotRepo.GetByID(t.Context(), rotationA.ID)
	require.NoError(t, err)
	require.NotNil(t, got.ScheduledMember)
	require.Equal(t, member.ID, got.ScheduledMember.ID)
	require.Equal(t, rotationA.ID, got.ScheduledMember.RotationID)
	require.Equal(t, 1, got.ScheduledMember.Order)
	require.Equal(t, user.ID, got.ScheduledMember.User.ID)
	require.Equal(t, "Alice Smith", got.ScheduledMember.User.Name)
	require.Equal(t, "alice@example.com", got.ScheduledMember.User.Email)
}
