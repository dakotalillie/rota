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

// fakes

type fakeTransactor struct{}

func (f *fakeTransactor) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type fakeRotationRepo struct {
	rotation *domain.Rotation
	err      error
}

func (f *fakeRotationRepo) GetByID(_ context.Context, _ string) (*domain.Rotation, error) {
	return f.rotation, f.err
}

type fakeUserRepo struct {
	getByIDUser *domain.User
	getByIDErr  error
	createUser  *domain.User
	createErr   error
}

func (f *fakeUserRepo) GetByID(_ context.Context, _ string) (*domain.User, error) {
	return f.getByIDUser, f.getByIDErr
}

func (f *fakeUserRepo) Create(_ context.Context, _, _ string) (*domain.User, error) {
	return f.createUser, f.createErr
}

type fakeMemberRepo struct {
	count        int
	countErr     error
	createMember *domain.Member
	createErr    error
}

func (f *fakeMemberRepo) CountByRotationID(_ context.Context, _ string) (int, error) {
	return f.count, f.countErr
}

func (f *fakeMemberRepo) Create(_ context.Context, _, _ string, _ int) (*domain.Member, error) {
	return f.createMember, f.createErr
}

// tests

func TestCreateMemberUseCase_Execute(t *testing.T) {
	existingRotation := &domain.Rotation{ID: "rot_01JQGF0000000000000000000", Name: "Platform On-Call"}
	existingUser := &domain.User{ID: "usr_01JQGF0000000000000000000", Name: "Alice", Email: "alice@example.com"}
	newUser := &domain.User{ID: "usr_01JQGF0000000000000000001", Name: "Bob", Email: "bob@example.com"}
	createdMember := &domain.Member{ID: "mem_01JQGF0000000000000000000", RotationID: existingRotation.ID, Order: 1}

	tests := []struct {
		name         string
		input        application.CreateMemberInput
		rotationRepo fakeRotationRepo
		userRepo     fakeUserRepo
		memberRepo   fakeMemberRepo
		wantMember   *domain.Member
		wantErr      error
	}{
		{
			name: "success - existing user",
			input: application.CreateMemberInput{
				RotationID: existingRotation.ID,
				UserID:     existingUser.ID,
			},
			rotationRepo: fakeRotationRepo{rotation: existingRotation},
			userRepo:     fakeUserRepo{getByIDUser: existingUser},
			memberRepo: fakeMemberRepo{
				count:        0,
				createMember: createdMember,
			},
			wantMember: &domain.Member{
				ID:         createdMember.ID,
				RotationID: createdMember.RotationID,
				Order:      createdMember.Order,
				User:       *existingUser,
			},
		},
		{
			name: "success - new user",
			input: application.CreateMemberInput{
				RotationID: existingRotation.ID,
				UserName:   "Bob",
				UserEmail:  "bob@example.com",
			},
			rotationRepo: fakeRotationRepo{rotation: existingRotation},
			userRepo:     fakeUserRepo{createUser: newUser},
			memberRepo: fakeMemberRepo{
				count:        0,
				createMember: createdMember,
			},
			wantMember: &domain.Member{
				ID:         createdMember.ID,
				RotationID: createdMember.RotationID,
				Order:      createdMember.Order,
				User:       *newUser,
			},
		},
		{
			name: "missing user fields - no user id, name, or email",
			input: application.CreateMemberInput{
				RotationID: existingRotation.ID,
			},
			wantErr: domain.ErrMissingUserFields,
		},
		{
			name: "missing user fields - no user id and only name provided",
			input: application.CreateMemberInput{
				RotationID: existingRotation.ID,
				UserName:   "Bob",
			},
			wantErr: domain.ErrMissingUserFields,
		},
		{
			name: "missing user fields - no user id and only email provided",
			input: application.CreateMemberInput{
				RotationID: existingRotation.ID,
				UserEmail:  "bob@example.com",
			},
			wantErr: domain.ErrMissingUserFields,
		},
		{
			name: "rotation not found",
			input: application.CreateMemberInput{
				RotationID: "rot_notfound",
				UserID:     existingUser.ID,
			},
			rotationRepo: fakeRotationRepo{err: domain.ErrRotationNotFound},
			wantErr:      domain.ErrRotationNotFound,
		},
		{
			name: "rotation full",
			input: application.CreateMemberInput{
				RotationID: existingRotation.ID,
				UserID:     existingUser.ID,
			},
			rotationRepo: fakeRotationRepo{rotation: existingRotation},
			memberRepo:   fakeMemberRepo{count: 20},
			wantErr:      domain.ErrRotationMembershipFull,
		},
		{
			name: "user not found",
			input: application.CreateMemberInput{
				RotationID: existingRotation.ID,
				UserID:     "usr_notfound",
			},
			rotationRepo: fakeRotationRepo{rotation: existingRotation},
			memberRepo:   fakeMemberRepo{count: 0},
			userRepo:     fakeUserRepo{getByIDErr: domain.ErrUserNotFound},
			wantErr:      domain.ErrUserNotFound,
		},
		{
			name: "duplicate member",
			input: application.CreateMemberInput{
				RotationID: existingRotation.ID,
				UserID:     existingUser.ID,
			},
			rotationRepo: fakeRotationRepo{rotation: existingRotation},
			memberRepo: fakeMemberRepo{
				count:     0,
				createErr: domain.ErrMemberAlreadyExists,
			},
			userRepo: fakeUserRepo{getByIDUser: existingUser},
			wantErr:  domain.ErrMemberAlreadyExists,
		},
		{
			name: "member repo count error propagates",
			input: application.CreateMemberInput{
				RotationID: existingRotation.ID,
				UserID:     existingUser.ID,
			},
			rotationRepo: fakeRotationRepo{rotation: existingRotation},
			memberRepo:   fakeMemberRepo{countErr: errors.New("db error")},
			wantErr:      errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := application.NewCreateMemberUseCase(
				&fakeTransactor{},
				&tc.rotationRepo,
				&tc.userRepo,
				&tc.memberRepo,
			)

			got, err := uc.Execute(context.Background(), tc.input)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, tc.wantErr.Error(), err.Error())
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.wantMember, got)
		})
	}
}
