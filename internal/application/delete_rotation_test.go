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

func TestDeleteRotationUseCase_Execute(t *testing.T) {
	const rotationID = "rot_01JQGF0000000000000000000"
	const memberID = "mem_01JQGF0000000000000000001"
	const userID = "usr_01JQGF0000000000000000002"

	rotationWithMember := &domain.Rotation{
		ID:   rotationID,
		Name: "Platform On-Call",
		Members: []domain.Member{
			{ID: memberID, RotationID: rotationID, User: domain.User{ID: userID}},
		},
	}
	rotationNoMembers := &domain.Rotation{
		ID:   rotationID,
		Name: "Platform On-Call",
	}

	tests := []struct {
		name                string
		input               application.DeleteRotationInput
		rotationRepo        fakeRotationRepo
		memberRepo          fakeMemberRepo
		overrideRepo        fakeOverrideRepo
		userRepo            fakeUserRepo
		wantErr             error
		wantRotationDeleted bool
		wantMemberDeleted   bool
		wantUserDeleted     bool
	}{
		{
			name:  "success - no members",
			input: application.DeleteRotationInput{RotationID: rotationID},
			rotationRepo: fakeRotationRepo{
				rotation: rotationNoMembers,
			},
			wantRotationDeleted: true,
		},
		{
			name:  "success - member's last membership, user deleted",
			input: application.DeleteRotationInput{RotationID: rotationID},
			rotationRepo: fakeRotationRepo{
				rotation: rotationWithMember,
			},
			userRepo: fakeUserRepo{
				countMembershipsCount: 0,
			},
			wantRotationDeleted: true,
			wantMemberDeleted:   true,
			wantUserDeleted:     true,
		},
		{
			name:  "success - member has other memberships, user retained",
			input: application.DeleteRotationInput{RotationID: rotationID},
			rotationRepo: fakeRotationRepo{
				rotation: rotationWithMember,
			},
			userRepo: fakeUserRepo{
				countMembershipsCount: 1,
			},
			wantRotationDeleted: true,
			wantMemberDeleted:   true,
			wantUserDeleted:     false,
		},
		{
			name:         "rotation not found",
			input:        application.DeleteRotationInput{RotationID: rotationID},
			rotationRepo: fakeRotationRepo{err: domain.ErrRotationNotFound},
			wantErr:      domain.ErrRotationNotFound,
		},
		{
			name:  "delete error propagates",
			input: application.DeleteRotationInput{RotationID: rotationID},
			rotationRepo: fakeRotationRepo{
				rotation:  rotationNoMembers,
				deleteErr: errors.New("db error"),
			},
			wantErr:             errors.New("db error"),
			wantRotationDeleted: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := application.NewDeleteRotationUseCase(
				&fakeTransactor{},
				&tc.rotationRepo,
				&tc.memberRepo,
				&tc.overrideRepo,
				&tc.userRepo,
			)

			err := uc.Execute(context.Background(), tc.input)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, tc.wantErr.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			if tc.wantRotationDeleted {
				require.Len(t, tc.rotationRepo.deleteCalls, 1)
				assert.Equal(t, rotationID, tc.rotationRepo.deleteCalls[0])
			} else {
				assert.Empty(t, tc.rotationRepo.deleteCalls)
			}

			if tc.wantMemberDeleted {
				require.Len(t, tc.memberRepo.deleteMemberCalls, 1)
				assert.Equal(t, memberID, tc.memberRepo.deleteMemberCalls[0])
			} else {
				assert.Empty(t, tc.memberRepo.deleteMemberCalls)
			}

			if tc.wantUserDeleted {
				require.Len(t, tc.userRepo.deleteUserCalls, 1)
				assert.Equal(t, userID, tc.userRepo.deleteUserCalls[0])
			} else {
				assert.Empty(t, tc.userRepo.deleteUserCalls)
			}
		})
	}
}
