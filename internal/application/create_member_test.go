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

func TestCreateMemberUseCase_Execute(t *testing.T) {
	existingRotation := &domain.Rotation{ID: "rot_01JQGF0000000000000000000", Name: "Platform On-Call"}
	existingUser := &domain.User{ID: "usr_01JQGF0000000000000000000", Name: "Alice", Email: "alice@example.com"}
	newUser := &domain.User{ID: "usr_01JQGF0000000000000000001", Name: "Bob", Email: "bob@example.com"}
	firstCreatedMember := &domain.Member{
		ID:         "mem_01JQGF0000000000000000000",
		RotationID: existingRotation.ID,
		Order:      1,
		Color:      domain.MemberColors[0],
	}
	secondCreatedMember := &domain.Member{
		ID:         "mem_01JQGF0000000000000000001",
		RotationID: existingRotation.ID,
		Order:      2,
		Color:      domain.MemberColors[1],
	}
	fixedNow := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name              string
		input             application.CreateMemberInput
		rotationRepo      fakeRotationRepo
		userRepo          fakeUserRepo
		memberRepo        fakeMemberRepo
		wantMember        *domain.Member
		wantErr           error
		wantCreateOrder   int
		wantCreateColor   string
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
				createMember: firstCreatedMember,
			},
			wantMember: &domain.Member{
				ID:         firstCreatedMember.ID,
				RotationID: firstCreatedMember.RotationID,
				Order:      firstCreatedMember.Order,
				Color:      firstCreatedMember.Color,
				User:       *existingUser,
			},
			wantCreateOrder:   1,
			wantCreateColor:   domain.MemberColors[0],
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
				createMember: secondCreatedMember,
			},
			wantMember: &domain.Member{
				ID:         secondCreatedMember.ID,
				RotationID: secondCreatedMember.RotationID,
				Order:      secondCreatedMember.Order,
				Color:      secondCreatedMember.Color,
				User:       *existingUser,
			},
			wantCreateOrder:   2,
			wantCreateColor:   domain.MemberColors[1],
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
				createMember: firstCreatedMember,
			},
			wantMember: &domain.Member{
				ID:         firstCreatedMember.ID,
				RotationID: firstCreatedMember.RotationID,
				Order:      firstCreatedMember.Order,
				Color:      firstCreatedMember.Color,
				User:       *newUser,
			},
			wantCreateOrder:   1,
			wantCreateColor:   domain.MemberColors[0],
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
				createMember: firstCreatedMember,
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
			require.Len(t, tc.memberRepo.createCalls, 1)
			assert.Equal(t, tc.input.RotationID, tc.memberRepo.createCalls[0].rotationID)
			assert.Equal(t, tc.wantCreateOrder, tc.memberRepo.createCalls[0].order)
			assert.Equal(t, tc.wantCreateColor, tc.memberRepo.createCalls[0].color)
			if tc.wantSetCurrCalled {
				require.Len(t, tc.memberRepo.setCurrCalls, 1)
				assert.Equal(t, tc.wantMember.ID, tc.memberRepo.setCurrCalls[0].memberID)
				assert.Equal(t, tc.input.Now, tc.memberRepo.setCurrCalls[0].at)
			} else {
				assert.Empty(t, tc.memberRepo.setCurrCalls)
			}
		})
	}
}
