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

func TestDeleteMemberUseCase_Execute(t *testing.T) {
	const rotationID = "rot_01JQGF0000000000000000000"
	fixedNow := time.Date(2026, 4, 6, 9, 0, 0, 0, time.UTC)

	alice := domain.Member{
		ID:         "mem_01JQGF0000000000000000001",
		RotationID: rotationID,
		Order:      1,
		User:       domain.User{ID: "usr_01JQGF0000000000000000001", Name: "Alice", Email: "alice@example.com"},
	}
	bob := domain.Member{
		ID:         "mem_01JQGF0000000000000000002",
		RotationID: rotationID,
		Order:      2,
		User:       domain.User{ID: "usr_01JQGF0000000000000000002", Name: "Bob", Email: "bob@example.com"},
	}
	charlie := domain.Member{
		ID:         "mem_01JQGF0000000000000000003",
		RotationID: rotationID,
		Order:      3,
		User:       domain.User{ID: "usr_01JQGF0000000000000000003", Name: "Charlie", Email: "charlie@example.com"},
	}

	rotation3 := &domain.Rotation{
		ID:            rotationID,
		Name:          "Platform On-Call",
		Members:       []domain.Member{alice, bob, charlie},
		CurrentMember: &alice,
	}

	tests := []struct {
		name                          string
		input                         application.DeleteMemberInput
		rotationRepo                  fakeRotationRepo
		memberRepo                    fakeMemberRepo
		overrideRepo                  fakeOverrideRepo
		userRepo                      fakeUserRepo
		wantErr                       error
		wantDeletedMember             string
		wantDeletedOverridesForMember string
		wantDeletedUser               string
		wantReorderCall               []string
		wantSetCurrMember             string
	}{
		{
			name:  "success - non-current member, user has other memberships",
			input: application.DeleteMemberInput{RotationID: rotationID, MemberID: bob.ID, Now: fixedNow},
			rotationRepo: fakeRotationRepo{rotation: &domain.Rotation{
				ID:            rotationID,
				Name:          "Platform On-Call",
				Members:       []domain.Member{alice, bob, charlie},
				CurrentMember: &alice,
			}},
			memberRepo: fakeMemberRepo{
				getByIDMember: &bob,
			},
			userRepo: fakeUserRepo{
				countMembershipsCount: 1, // still has another membership
			},
			wantDeletedMember:             bob.ID,
			wantDeletedOverridesForMember: bob.ID,
			wantReorderCall:               []string{alice.ID, charlie.ID},
		},
		{
			name:  "success - user's last membership, user deleted",
			input: application.DeleteMemberInput{RotationID: rotationID, MemberID: bob.ID, Now: fixedNow},
			rotationRepo: fakeRotationRepo{rotation: &domain.Rotation{
				ID:            rotationID,
				Name:          "Platform On-Call",
				Members:       []domain.Member{alice, bob, charlie},
				CurrentMember: &alice,
			}},
			memberRepo: fakeMemberRepo{
				getByIDMember: &bob,
			},
			userRepo: fakeUserRepo{
				countMembershipsCount: 0,
			},
			wantDeletedMember:             bob.ID,
			wantDeletedOverridesForMember: bob.ID,
			wantDeletedUser:               bob.User.ID,
			wantReorderCall:               []string{alice.ID, charlie.ID},
		},
		{
			name:         "success - current member deleted, next member promoted",
			input:        application.DeleteMemberInput{RotationID: rotationID, MemberID: alice.ID, Now: fixedNow},
			rotationRepo: fakeRotationRepo{rotation: rotation3},
			memberRepo: fakeMemberRepo{
				getByIDMember: &alice,
			},
			userRepo: fakeUserRepo{
				countMembershipsCount: 0,
			},
			wantDeletedMember:             alice.ID,
			wantDeletedOverridesForMember: alice.ID,
			wantDeletedUser:               alice.User.ID,
			wantReorderCall:               []string{bob.ID, charlie.ID},
			wantSetCurrMember:             bob.ID, // alice was index 0, next = index 0 % 2 = bob
		},
		{
			name:  "success - current member deleted, wrap-around to first",
			input: application.DeleteMemberInput{RotationID: rotationID, MemberID: charlie.ID, Now: fixedNow},
			rotationRepo: fakeRotationRepo{rotation: &domain.Rotation{
				ID:            rotationID,
				Name:          "Platform On-Call",
				Members:       []domain.Member{alice, bob, charlie},
				CurrentMember: &charlie,
			}},
			memberRepo: fakeMemberRepo{
				getByIDMember: &charlie,
			},
			userRepo: fakeUserRepo{
				countMembershipsCount: 0,
			},
			wantDeletedMember:             charlie.ID,
			wantDeletedOverridesForMember: charlie.ID,
			wantDeletedUser:               charlie.User.ID,
			wantReorderCall:               []string{alice.ID, bob.ID},
			wantSetCurrMember:             alice.ID, // charlie was index 2, next = 2 % 2 = alice
		},
		{
			name:  "success - last member deleted",
			input: application.DeleteMemberInput{RotationID: rotationID, MemberID: alice.ID, Now: fixedNow},
			rotationRepo: fakeRotationRepo{rotation: &domain.Rotation{
				ID:            rotationID,
				Name:          "Platform On-Call",
				Members:       []domain.Member{alice},
				CurrentMember: &alice,
			}},
			memberRepo: fakeMemberRepo{
				getByIDMember: &alice,
			},
			userRepo: fakeUserRepo{
				countMembershipsCount: 0,
			},
			wantDeletedMember:             alice.ID,
			wantDeletedOverridesForMember: alice.ID,
			wantDeletedUser:               alice.User.ID,
			// no reorder, no SetCurrentMember
		},
		{
			name:         "rotation not found",
			input:        application.DeleteMemberInput{RotationID: "rot_notfound", MemberID: alice.ID},
			rotationRepo: fakeRotationRepo{err: domain.ErrRotationNotFound},
			wantErr:      domain.ErrRotationNotFound,
		},
		{
			name:         "member not found",
			input:        application.DeleteMemberInput{RotationID: rotationID, MemberID: "mem_notfound"},
			rotationRepo: fakeRotationRepo{rotation: rotation3},
			memberRepo:   fakeMemberRepo{getByIDErr: domain.ErrMemberNotFound},
			wantErr:      domain.ErrMemberNotFound,
		},
		{
			name:         "override delete error propagates",
			input:        application.DeleteMemberInput{RotationID: rotationID, MemberID: bob.ID},
			rotationRepo: fakeRotationRepo{rotation: rotation3},
			memberRepo:   fakeMemberRepo{getByIDMember: &bob},
			overrideRepo: fakeOverrideRepo{deleteByMemberIDErr: errors.New("db error")},
			wantErr:      errors.New("db error"),
		},
		{
			name:         "member delete error propagates",
			input:        application.DeleteMemberInput{RotationID: rotationID, MemberID: bob.ID},
			rotationRepo: fakeRotationRepo{rotation: rotation3},
			memberRepo:   fakeMemberRepo{getByIDMember: &bob, deleteMemberErr: errors.New("db error")},
			wantErr:      errors.New("db error"),
		},
		{
			name:         "count memberships error propagates",
			input:        application.DeleteMemberInput{RotationID: rotationID, MemberID: bob.ID},
			rotationRepo: fakeRotationRepo{rotation: rotation3},
			memberRepo:   fakeMemberRepo{getByIDMember: &bob},
			userRepo:     fakeUserRepo{countMembershipsErr: errors.New("db error")},
			wantErr:      errors.New("db error"),
		},
		{
			name:         "user delete error propagates",
			input:        application.DeleteMemberInput{RotationID: rotationID, MemberID: bob.ID},
			rotationRepo: fakeRotationRepo{rotation: rotation3},
			memberRepo:   fakeMemberRepo{getByIDMember: &bob},
			userRepo:     fakeUserRepo{countMembershipsCount: 0, deleteUserErr: errors.New("db error")},
			wantErr:      errors.New("db error"),
		},
		{
			name:  "reorder error propagates",
			input: application.DeleteMemberInput{RotationID: rotationID, MemberID: bob.ID},
			rotationRepo: fakeRotationRepo{rotation: &domain.Rotation{
				ID:      rotationID,
				Members: []domain.Member{alice, bob, charlie},
			}},
			memberRepo: fakeMemberRepo{getByIDMember: &bob, reorderErr: errors.New("db error")},
			wantErr:    errors.New("db error"),
		},
		{
			name:         "set current member error propagates",
			input:        application.DeleteMemberInput{RotationID: rotationID, MemberID: alice.ID, Now: fixedNow},
			rotationRepo: fakeRotationRepo{rotation: rotation3},
			memberRepo:   fakeMemberRepo{getByIDMember: &alice, setCurrErr: errors.New("db error")},
			wantErr:      errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := application.NewDeleteMemberUseCase(
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
				return
			}

			require.NoError(t, err)

			if tc.wantDeletedMember != "" {
				require.Len(t, tc.memberRepo.deleteMemberCalls, 1)
				assert.Equal(t, tc.wantDeletedMember, tc.memberRepo.deleteMemberCalls[0])
			} else {
				assert.Empty(t, tc.memberRepo.deleteMemberCalls)
			}

			if tc.wantDeletedOverridesForMember != "" {
				require.Len(t, tc.overrideRepo.deleteByMemberIDCalls, 1)
				assert.Equal(t, tc.wantDeletedOverridesForMember, tc.overrideRepo.deleteByMemberIDCalls[0])
			} else {
				assert.Empty(t, tc.overrideRepo.deleteByMemberIDCalls)
			}

			if tc.wantDeletedUser != "" {
				require.Len(t, tc.userRepo.deleteUserCalls, 1)
				assert.Equal(t, tc.wantDeletedUser, tc.userRepo.deleteUserCalls[0])
			} else {
				assert.Empty(t, tc.userRepo.deleteUserCalls)
			}

			if tc.wantReorderCall != nil {
				require.Len(t, tc.memberRepo.reorderCalls, 1)
				assert.Equal(t, tc.wantReorderCall, tc.memberRepo.reorderCalls[0])
			} else {
				assert.Empty(t, tc.memberRepo.reorderCalls)
			}

			if tc.wantSetCurrMember != "" {
				require.Len(t, tc.memberRepo.setCurrCalls, 1)
				assert.Equal(t, tc.wantSetCurrMember, tc.memberRepo.setCurrCalls[0].memberID)
				assert.Equal(t, tc.input.Now, tc.memberRepo.setCurrCalls[0].at)
			} else {
				assert.Empty(t, tc.memberRepo.setCurrCalls)
			}
		})
	}
}
