package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

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
	count     int
	countErr  error
	rotation  *domain.Rotation
	rotations []*domain.Rotation
	err       error
}

func (f *fakeRotationRepo) Count(_ context.Context) (int, error) {
	return f.count, f.countErr
}

func (f *fakeRotationRepo) Create(_ context.Context, rot *domain.Rotation) (*domain.Rotation, error) {
	return rot, f.err
}

func (f *fakeRotationRepo) GetByID(_ context.Context, _ string) (*domain.Rotation, error) {
	return f.rotation, f.err
}

func (f *fakeRotationRepo) List(_ context.Context) ([]*domain.Rotation, error) {
	return f.rotations, f.err
}

type fakeUserRepo struct {
	getByIDUser           *domain.User
	getByIDErr            error
	createUser            *domain.User
	createErr             error
	countMembershipsCount int
	countMembershipsErr   error
	deleteUserCalls       []string
	deleteUserErr         error
}

func (f *fakeUserRepo) GetByID(_ context.Context, _ string) (*domain.User, error) {
	return f.getByIDUser, f.getByIDErr
}

func (f *fakeUserRepo) Create(_ context.Context, _, _ string) (*domain.User, error) {
	return f.createUser, f.createErr
}

func (f *fakeUserRepo) CountMemberships(_ context.Context, _ string) (int, error) {
	return f.countMembershipsCount, f.countMembershipsErr
}

func (f *fakeUserRepo) Delete(_ context.Context, userID string) error {
	f.deleteUserCalls = append(f.deleteUserCalls, userID)
	return f.deleteUserErr
}

type fakeMemberRepo struct {
	count             int
	countErr          error
	createMember      *domain.Member
	createErr         error
	getByIDMember     *domain.Member
	getByIDErr        error
	deleteMemberCalls []string
	deleteMemberErr   error
	setCurrErr        error
	setCurrCalls      []struct {
		memberID string
		at       time.Time
	}
	reorderErr   error
	reorderCalls [][]string
}

func (f *fakeMemberRepo) CountByRotationID(_ context.Context, _ string) (int, error) {
	return f.count, f.countErr
}

func (f *fakeMemberRepo) Create(_ context.Context, _, _ string, _ int) (*domain.Member, error) {
	return f.createMember, f.createErr
}

func (f *fakeMemberRepo) GetByID(_ context.Context, _, _ string) (*domain.Member, error) {
	return f.getByIDMember, f.getByIDErr
}

func (f *fakeMemberRepo) Delete(_ context.Context, memberID string) error {
	f.deleteMemberCalls = append(f.deleteMemberCalls, memberID)
	return f.deleteMemberErr
}

func (f *fakeMemberRepo) SetCurrentMember(_ context.Context, _ string, memberID string, at time.Time) error {
	f.setCurrCalls = append(f.setCurrCalls, struct {
		memberID string
		at       time.Time
	}{memberID, at})
	return f.setCurrErr
}

func (f *fakeMemberRepo) ReorderMembers(_ context.Context, _ string, memberIDs []string) error {
	f.reorderCalls = append(f.reorderCalls, memberIDs)
	return f.reorderErr
}

// tests

func TestCreateMemberUseCase_Execute(t *testing.T) {
	existingRotation := &domain.Rotation{ID: "rot_01JQGF0000000000000000000", Name: "Platform On-Call"}
	existingUser := &domain.User{ID: "usr_01JQGF0000000000000000000", Name: "Alice", Email: "alice@example.com"}
	newUser := &domain.User{ID: "usr_01JQGF0000000000000000001", Name: "Bob", Email: "bob@example.com"}
	createdMember := &domain.Member{ID: "mem_01JQGF0000000000000000000", RotationID: existingRotation.ID, Order: 1}
	fixedNow := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name              string
		input             application.CreateMemberInput
		rotationRepo      fakeRotationRepo
		userRepo          fakeUserRepo
		memberRepo        fakeMemberRepo
		wantMember        *domain.Member
		wantErr           error
		wantSetCurrCalled bool
	}{
		{
			name: "success - first member becomes current",
			input: application.CreateMemberInput{
				RotationID: existingRotation.ID,
				UserID:     existingUser.ID,
				Now:        fixedNow,
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
			wantSetCurrCalled: true,
		},
		{
			name: "success - subsequent member does not become current",
			input: application.CreateMemberInput{
				RotationID: existingRotation.ID,
				UserID:     existingUser.ID,
				Now:        fixedNow,
			},
			rotationRepo: fakeRotationRepo{rotation: existingRotation},
			userRepo:     fakeUserRepo{getByIDUser: existingUser},
			memberRepo: fakeMemberRepo{
				count:        1,
				createMember: createdMember,
			},
			wantMember: &domain.Member{
				ID:         createdMember.ID,
				RotationID: createdMember.RotationID,
				Order:      createdMember.Order,
				User:       *existingUser,
			},
			wantSetCurrCalled: false,
		},
		{
			name: "success - new user",
			input: application.CreateMemberInput{
				RotationID: existingRotation.ID,
				UserName:   "Bob",
				UserEmail:  "bob@example.com",
				Now:        fixedNow,
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
			wantSetCurrCalled: true,
		},
		{
			name: "set current member error propagates",
			input: application.CreateMemberInput{
				RotationID: existingRotation.ID,
				UserID:     existingUser.ID,
				Now:        fixedNow,
			},
			rotationRepo: fakeRotationRepo{rotation: existingRotation},
			userRepo:     fakeUserRepo{getByIDUser: existingUser},
			memberRepo: fakeMemberRepo{
				count:        0,
				createMember: createdMember,
				setCurrErr:   errors.New("db error"),
			},
			wantErr: errors.New("db error"),
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
			if tc.wantSetCurrCalled {
				require.Len(t, tc.memberRepo.setCurrCalls, 1)
				assert.Equal(t, createdMember.ID, tc.memberRepo.setCurrCalls[0].memberID)
				assert.Equal(t, tc.input.Now, tc.memberRepo.setCurrCalls[0].at)
			} else {
				assert.Empty(t, tc.memberRepo.setCurrCalls)
			}
		})
	}
}
