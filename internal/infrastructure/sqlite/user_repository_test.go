package sqlite_test

import (
	"testing"

	"github.com/dakotalillie/rota/internal/domain"
	"github.com/dakotalillie/rota/internal/infrastructure/sqlite"
	"github.com/stretchr/testify/require"
)

var userA = &domain.User{
	ID:    "usr_01JQGF0000000000000000000",
	Name:  "Alice Smith",
	Email: "alice@example.com",
}

func TestUserRepository_GetByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db := openTestDB(t)
		userRepo := sqlite.NewUserRepository(db)

		created, err := userRepo.Create(t.Context(), userA.Name, userA.Email)
		require.NoError(t, err)

		found, err := userRepo.GetByID(t.Context(), created.ID)
		require.NoError(t, err)
		require.Equal(t, created.ID, found.ID)
		require.Equal(t, userA.Name, found.Name)
		require.Equal(t, userA.Email, found.Email)
	})

	t.Run("not found", func(t *testing.T) {
		db := openTestDB(t)
		userRepo := sqlite.NewUserRepository(db)

		_, err := userRepo.GetByID(t.Context(), "usr_99999999999999999999999999")
		require.ErrorIs(t, err, domain.ErrUserNotFound)
	})
}

func TestUserRepository_CountMemberships(t *testing.T) {
	t.Run("zero memberships", func(t *testing.T) {
		db := openTestDB(t)
		userRepo := sqlite.NewUserRepository(db)

		user, err := userRepo.Create(t.Context(), userA.Name, userA.Email)
		require.NoError(t, err)

		count, err := userRepo.CountMemberships(t.Context(), user.ID)
		require.NoError(t, err)
		require.Equal(t, 0, count)
	})

	t.Run("one membership", func(t *testing.T) {
		db := openTestDB(t)
		rotRepo := sqlite.NewRotationRepository(db)
		userRepo := sqlite.NewUserRepository(db)
		memberRepo := sqlite.NewMemberRepository(db)

		require.NoError(t, rotRepo.UpsertRotation(t.Context(), rotationA))
		user, err := userRepo.Create(t.Context(), userA.Name, userA.Email)
		require.NoError(t, err)
		_, err = memberRepo.Create(t.Context(), rotationA.ID, user.ID, 1)
		require.NoError(t, err)

		count, err := userRepo.CountMemberships(t.Context(), user.ID)
		require.NoError(t, err)
		require.Equal(t, 1, count)
	})

	t.Run("two memberships across rotations", func(t *testing.T) {
		db := openTestDB(t)
		rotRepo := sqlite.NewRotationRepository(db)
		userRepo := sqlite.NewUserRepository(db)
		memberRepo := sqlite.NewMemberRepository(db)

		rotB := &domain.Rotation{ID: "rot_01JQGF0000000000000000001", Name: "Second Rotation"}
		require.NoError(t, rotRepo.UpsertRotation(t.Context(), rotationA))
		require.NoError(t, rotRepo.UpsertRotation(t.Context(), rotB))
		user, err := userRepo.Create(t.Context(), userA.Name, userA.Email)
		require.NoError(t, err)
		_, err = memberRepo.Create(t.Context(), rotationA.ID, user.ID, 1)
		require.NoError(t, err)
		_, err = memberRepo.Create(t.Context(), rotB.ID, user.ID, 1)
		require.NoError(t, err)

		count, err := userRepo.CountMemberships(t.Context(), user.ID)
		require.NoError(t, err)
		require.Equal(t, 2, count)
	})
}

func TestUserRepository_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db := openTestDB(t)
		userRepo := sqlite.NewUserRepository(db)

		user, err := userRepo.Create(t.Context(), userA.Name, userA.Email)
		require.NoError(t, err)

		require.NoError(t, userRepo.Delete(t.Context(), user.ID))

		_, err = userRepo.GetByID(t.Context(), user.ID)
		require.ErrorIs(t, err, domain.ErrUserNotFound)
	})
}

func TestUserRepository_Create(t *testing.T) {
	t.Run("creates new user", func(t *testing.T) {
		db := openTestDB(t)
		userRepo := sqlite.NewUserRepository(db)

		user, err := userRepo.Create(t.Context(), userA.Name, userA.Email)
		require.NoError(t, err)
		require.NotEmpty(t, user.ID)
		require.Equal(t, userA.Name, user.Name)
		require.Equal(t, userA.Email, user.Email)
	})

	t.Run("same email returns existing user", func(t *testing.T) {
		// Calling Create twice with the same email should return the same user ID,
		// not create a duplicate.
		db := openTestDB(t)
		userRepo := sqlite.NewUserRepository(db)

		u1, err := userRepo.Create(t.Context(), userA.Name, userA.Email)
		require.NoError(t, err)

		u2, err := userRepo.Create(t.Context(), userA.Name, userA.Email)
		require.NoError(t, err)

		require.Equal(t, u1.ID, u2.ID)
	})

	t.Run("same email across different rotations reuses user", func(t *testing.T) {
		// Adding the same email as a member of two different rotations should
		// resolve to the same user ID.
		db := openTestDB(t)
		rotRepo := sqlite.NewRotationRepository(db)
		userRepo := sqlite.NewUserRepository(db)
		memberRepo := sqlite.NewMemberRepository(db)

		rotB := &domain.Rotation{ID: "rot_01JQGF0000000000000000001", Name: "Second Rotation"}
		require.NoError(t, rotRepo.UpsertRotation(t.Context(), rotationA))
		require.NoError(t, rotRepo.UpsertRotation(t.Context(), rotB))

		u1, err := userRepo.Create(t.Context(), "Alice", userA.Email)
		require.NoError(t, err)
		_, err = memberRepo.Create(t.Context(), rotationA.ID, u1.ID, 1)
		require.NoError(t, err)

		u2, err := userRepo.Create(t.Context(), "Alice", userA.Email)
		require.NoError(t, err)
		_, err = memberRepo.Create(t.Context(), rotB.ID, u2.ID, 1)
		require.NoError(t, err)

		require.Equal(t, u1.ID, u2.ID, "same email should resolve to same user ID")
	})
}
