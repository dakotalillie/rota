package sqlite_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/dakotalillie/rota/internal/domain"
	"github.com/dakotalillie/rota/internal/infrastructure/sqlite"
	"github.com/stretchr/testify/require"
)

func TestMemberRepository_CountByRotationID(t *testing.T) {
	tests := []struct {
		name       string
		seedCount  int
		rotationID string
		wantCount  int
	}{
		{
			name:       "empty",
			seedCount:  0,
			rotationID: rotationA.ID,
			wantCount:  0,
		},
		{
			name:       "one member",
			seedCount:  1,
			rotationID: rotationA.ID,
			wantCount:  1,
		},
		{
			name:       "multiple members",
			seedCount:  3,
			rotationID: rotationA.ID,
			wantCount:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openTestDB(t)
			rotRepo := sqlite.NewRotationRepository(db)
			userRepo := sqlite.NewUserRepository(db)
			memberRepo := sqlite.NewMemberRepository(db)

			require.NoError(t, rotRepo.UpsertRotation(t.Context(), rotationA))
			for i := range tt.seedCount {
				user, err := userRepo.Create(t.Context(), "User", fmt.Sprintf("user%d@example.com", i))
				require.NoError(t, err)
				_, err = memberRepo.Create(t.Context(), rotationA.ID, user.ID, i+1, domain.MemberColors[i%len(domain.MemberColors)])
				require.NoError(t, err)
			}

			count, err := memberRepo.CountByRotationID(t.Context(), tt.rotationID)
			require.NoError(t, err)
			require.Equal(t, tt.wantCount, count)
		})
	}
}

func TestMemberRepository_CreateMember(t *testing.T) {
	tests := []struct {
		name         string
		position     int
		color        string
		wantPosition int
	}{
		{
			name:         "success - first member",
			position:     1,
			color:        domain.MemberColors[0],
			wantPosition: 1,
		},
		{
			name:         "success - second member gets position 2",
			position:     2,
			color:        domain.MemberColors[1],
			wantPosition: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openTestDB(t)
			rotRepo := sqlite.NewRotationRepository(db)
			userRepo := sqlite.NewUserRepository(db)
			memberRepo := sqlite.NewMemberRepository(db)

			require.NoError(t, rotRepo.UpsertRotation(t.Context(), rotationA))
			user, err := userRepo.Create(t.Context(), "Alice Smith", "alice@example.com")
			require.NoError(t, err)

			member, err := memberRepo.Create(t.Context(), rotationA.ID, user.ID, tt.position, tt.color)
			require.NoError(t, err)
			require.NotEmpty(t, member.ID)
			require.Equal(t, rotationA.ID, member.RotationID)
			require.Equal(t, user.ID, member.User.ID)
			require.Equal(t, tt.wantPosition, member.Position)
			require.Equal(t, tt.color, member.Color)
		})
	}
}

func TestMemberRepository_CreateMember_DuplicateUser(t *testing.T) {
	db := openTestDB(t)
	rotRepo := sqlite.NewRotationRepository(db)
	userRepo := sqlite.NewUserRepository(db)
	memberRepo := sqlite.NewMemberRepository(db)

	require.NoError(t, rotRepo.UpsertRotation(t.Context(), rotationA))
	user, err := userRepo.Create(t.Context(), "Alice", "alice@example.com")
	require.NoError(t, err)

	_, err = memberRepo.Create(t.Context(), rotationA.ID, user.ID, 1, domain.MemberColors[0])
	require.NoError(t, err)

	_, err = memberRepo.Create(t.Context(), rotationA.ID, user.ID, 2, domain.MemberColors[1])
	require.ErrorIs(t, err, domain.ErrMemberAlreadyExists)
}

func TestMemberRepository_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db := openTestDB(t)
		rotRepo := sqlite.NewRotationRepository(db)
		userRepo := sqlite.NewUserRepository(db)
		memberRepo := sqlite.NewMemberRepository(db)

		require.NoError(t, rotRepo.UpsertRotation(t.Context(), rotationA))
		user, err := userRepo.Create(t.Context(), "Alice", "alice@example.com")
		require.NoError(t, err)
		created, err := memberRepo.Create(t.Context(), rotationA.ID, user.ID, 1, domain.MemberColors[0])
		require.NoError(t, err)

		found, err := memberRepo.GetByID(t.Context(), rotationA.ID, created.ID)
		require.NoError(t, err)
		require.Equal(t, created.ID, found.ID)
		require.Equal(t, rotationA.ID, found.RotationID)
		require.Equal(t, user.ID, found.User.ID)
		require.Equal(t, 1, found.Position)
		require.Equal(t, domain.MemberColors[0], found.Color)
	})

	t.Run("not found", func(t *testing.T) {
		db := openTestDB(t)
		rotRepo := sqlite.NewRotationRepository(db)
		memberRepo := sqlite.NewMemberRepository(db)

		require.NoError(t, rotRepo.UpsertRotation(t.Context(), rotationA))

		_, err := memberRepo.GetByID(t.Context(), rotationA.ID, "mem_99999999999999999999999999")
		require.ErrorIs(t, err, domain.ErrMemberNotFound)
	})

	t.Run("wrong rotation", func(t *testing.T) {
		db := openTestDB(t)
		rotRepo := sqlite.NewRotationRepository(db)
		userRepo := sqlite.NewUserRepository(db)
		memberRepo := sqlite.NewMemberRepository(db)

		rotB := &domain.Rotation{ID: "rot_01JQGF0000000000000000001", Name: "Other Rotation"}
		require.NoError(t, rotRepo.UpsertRotation(t.Context(), rotationA))
		require.NoError(t, rotRepo.UpsertRotation(t.Context(), rotB))
		user, err := userRepo.Create(t.Context(), "Alice", "alice@example.com")
		require.NoError(t, err)
		created, err := memberRepo.Create(t.Context(), rotationA.ID, user.ID, 1, domain.MemberColors[0])
		require.NoError(t, err)

		// Looking up the member using a different rotation ID should fail.
		_, err = memberRepo.GetByID(t.Context(), rotB.ID, created.ID)
		require.ErrorIs(t, err, domain.ErrMemberNotFound)
	})
}

func TestMemberRepository_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db := openTestDB(t)
		rotRepo := sqlite.NewRotationRepository(db)
		userRepo := sqlite.NewUserRepository(db)
		memberRepo := sqlite.NewMemberRepository(db)

		require.NoError(t, rotRepo.UpsertRotation(t.Context(), rotationA))
		user, err := userRepo.Create(t.Context(), "Alice", "alice@example.com")
		require.NoError(t, err)
		created, err := memberRepo.Create(t.Context(), rotationA.ID, user.ID, 1, domain.MemberColors[0])
		require.NoError(t, err)

		require.NoError(t, memberRepo.Delete(t.Context(), created.ID))

		count, err := memberRepo.CountByRotationID(t.Context(), rotationA.ID)
		require.NoError(t, err)
		require.Equal(t, 0, count)
	})
}

func TestMemberRepository_ReorderMembers(t *testing.T) {
	seedMembers := func(t *testing.T, db *sql.DB, count int) (rotRepo *sqlite.RotationRepository, memberRepo *sqlite.MemberRepository, memberIDs []string) {
		t.Helper()
		rotRepo = sqlite.NewRotationRepository(db)
		userRepo := sqlite.NewUserRepository(db)
		memberRepo = sqlite.NewMemberRepository(db)

		require.NoError(t, rotRepo.UpsertRotation(t.Context(), rotationA))
		memberIDs = make([]string, count)
		for i := range count {
			user, err := userRepo.Create(t.Context(), fmt.Sprintf("User%d", i+1), fmt.Sprintf("user%d@example.com", i+1))
			require.NoError(t, err)
			m, err := memberRepo.Create(t.Context(), rotationA.ID, user.ID, i+1, domain.MemberColors[i%len(domain.MemberColors)])
			require.NoError(t, err)
			memberIDs[i] = m.ID
		}
		return rotRepo, memberRepo, memberIDs
	}

	t.Run("reorders members and persists new position", func(t *testing.T) {
		db := openTestDB(t)
		rotRepo, memberRepo, ids := seedMembers(t, db, 3)

		// Reverse the position: [1,2,3] → [3,2,1]
		reversed := []string{ids[2], ids[1], ids[0]}
		require.NoError(t, memberRepo.ReorderMembers(t.Context(), rotationA.ID, reversed))

		rotation, err := rotRepo.GetByID(t.Context(), rotationA.ID)
		require.NoError(t, err)

		positionByID := make(map[string]int, len(rotation.Members))
		colorByID := make(map[string]string, len(rotation.Members))
		for _, m := range rotation.Members {
			positionByID[m.ID] = m.Position
			colorByID[m.ID] = m.Color
		}
		require.Equal(t, 1, positionByID[ids[2]], "ids[2] should now be position 1")
		require.Equal(t, 2, positionByID[ids[1]], "ids[1] should remain position 2")
		require.Equal(t, 3, positionByID[ids[0]], "ids[0] should now be position 3")
		require.Equal(t, domain.MemberColors[2], colorByID[ids[2]])
		require.Equal(t, domain.MemberColors[1], colorByID[ids[1]])
		require.Equal(t, domain.MemberColors[0], colorByID[ids[0]])
	})

	t.Run("single member - no-op", func(t *testing.T) {
		db := openTestDB(t)
		_, memberRepo, ids := seedMembers(t, db, 1)

		require.NoError(t, memberRepo.ReorderMembers(t.Context(), rotationA.ID, ids))
	})
}
