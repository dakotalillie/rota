package sqlite_test

import (
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
				_, err = memberRepo.Create(t.Context(), rotationA.ID, user.ID, i+1)
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
		name      string
		order     int
		wantOrder int
	}{
		{
			name:      "success - first member",
			order:     1,
			wantOrder: 1,
		},
		{
			name:      "success - second member gets order 2",
			order:     2,
			wantOrder: 2,
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

			member, err := memberRepo.Create(t.Context(), rotationA.ID, user.ID, tt.order)
			require.NoError(t, err)
			require.NotEmpty(t, member.ID)
			require.Equal(t, rotationA.ID, member.RotationID)
			require.Equal(t, user.ID, member.User.ID)
			require.Equal(t, tt.wantOrder, member.Order)
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

	_, err = memberRepo.Create(t.Context(), rotationA.ID, user.ID, 1)
	require.NoError(t, err)

	_, err = memberRepo.Create(t.Context(), rotationA.ID, user.ID, 2)
	require.ErrorIs(t, err, domain.ErrMemberAlreadyExists)
}
