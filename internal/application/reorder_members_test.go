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

// fakeRotationRepoSequenced returns successive results from a queue of
// (rotation, error) pairs, one per GetByID call. All other methods delegate
// to a fakeRotationRepo.
type fakeRotationRepoSequenced struct {
	results []struct {
		rotation *domain.Rotation
		err      error
	}
	idx int
}

func (f *fakeRotationRepoSequenced) Count(_ context.Context) (int, error) { return 0, nil }

func (f *fakeRotationRepoSequenced) Create(_ context.Context, rot *domain.Rotation) (*domain.Rotation, error) {
	return rot, nil
}

func (f *fakeRotationRepoSequenced) GetByID(_ context.Context, _ string) (*domain.Rotation, error) {
	if f.idx >= len(f.results) {
		return nil, errors.New("unexpected extra GetByID call")
	}
	r := f.results[f.idx]
	f.idx++
	return r.rotation, r.err
}

func (f *fakeRotationRepoSequenced) List(_ context.Context) ([]*domain.Rotation, error) {
	return nil, nil
}

func TestReorderMembersUseCase_Execute(t *testing.T) {
	const rotationID = "rot_01JQGF0000000000000000000"

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

	rotation := &domain.Rotation{
		ID:      rotationID,
		Name:    "Platform On-Call",
		Members: []domain.Member{alice, bob, charlie},
	}

	tests := []struct {
		name           string
		rotationRepo   domain.RotationRepository
		memberRepo     fakeMemberRepo
		input          application.ReorderMembersInput
		wantErr        error
		wantMemberIDs  []string // ordered IDs expected in returned rotation
		wantReorderArg []string // IDs passed to ReorderMembers
	}{
		{
			name: "success - reorders members",
			rotationRepo: &fakeRotationRepoSequenced{results: []struct {
				rotation *domain.Rotation
				err      error
			}{
				{rotation: rotation},
				{rotation: &domain.Rotation{
					ID:   rotationID,
					Name: "Platform On-Call",
					Members: []domain.Member{
						{ID: charlie.ID, Order: 1, User: charlie.User},
						{ID: alice.ID, Order: 2, User: alice.User},
						{ID: bob.ID, Order: 3, User: bob.User},
					},
				}},
			}},
			input: application.ReorderMembersInput{
				RotationID: rotationID,
				MemberIDs:  []string{charlie.ID, alice.ID, bob.ID},
			},
			wantMemberIDs:  []string{charlie.ID, alice.ID, bob.ID},
			wantReorderArg: []string{charlie.ID, alice.ID, bob.ID},
		},
		{
			name:         "rotation not found",
			rotationRepo: &fakeRotationRepo{err: domain.ErrRotationNotFound},
			input: application.ReorderMembersInput{
				RotationID: "rot_notfound",
				MemberIDs:  []string{alice.ID, bob.ID, charlie.ID},
			},
			wantErr: domain.ErrRotationNotFound,
		},
		{
			name:         "member count mismatch - too few",
			rotationRepo: &fakeRotationRepo{rotation: rotation},
			input: application.ReorderMembersInput{
				RotationID: rotationID,
				MemberIDs:  []string{alice.ID, bob.ID},
			},
			wantErr: domain.ErrMemberMismatch,
		},
		{
			name:         "member count mismatch - too many",
			rotationRepo: &fakeRotationRepo{rotation: rotation},
			input: application.ReorderMembersInput{
				RotationID: rotationID,
				MemberIDs:  []string{alice.ID, bob.ID, charlie.ID, "mem_extra"},
			},
			wantErr: domain.ErrMemberMismatch,
		},
		{
			name:         "unknown member ID",
			rotationRepo: &fakeRotationRepo{rotation: rotation},
			input: application.ReorderMembersInput{
				RotationID: rotationID,
				MemberIDs:  []string{alice.ID, bob.ID, "mem_unknown"},
			},
			wantErr: domain.ErrMemberMismatch,
		},
		{
			name:         "duplicate member ID",
			rotationRepo: &fakeRotationRepo{rotation: rotation},
			input: application.ReorderMembersInput{
				RotationID: rotationID,
				MemberIDs:  []string{alice.ID, alice.ID, charlie.ID},
			},
			wantErr: domain.ErrMemberMismatch,
		},
		{
			name:         "reorder repo error propagates",
			rotationRepo: &fakeRotationRepo{rotation: rotation},
			memberRepo:   fakeMemberRepo{reorderErr: errors.New("db error")},
			input: application.ReorderMembersInput{
				RotationID: rotationID,
				MemberIDs:  []string{alice.ID, bob.ID, charlie.ID},
			},
			wantErr: errors.New("db error"),
		},
		{
			name: "reload error propagates",
			rotationRepo: &fakeRotationRepoSequenced{results: []struct {
				rotation *domain.Rotation
				err      error
			}{
				{rotation: rotation},
				{err: errors.New("db error")},
			}},
			input: application.ReorderMembersInput{
				RotationID: rotationID,
				MemberIDs:  []string{alice.ID, bob.ID, charlie.ID},
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rotationRepo := tc.rotationRepo
			if rotationRepo == nil {
				rotationRepo = &fakeRotationRepo{rotation: rotation}
			}

			uc := application.NewReorderMembersUseCase(
				&fakeTransactor{},
				rotationRepo,
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
			require.NotNil(t, got)

			gotIDs := make([]string, len(got.Members))
			for i, m := range got.Members {
				gotIDs[i] = m.ID
			}
			assert.Equal(t, tc.wantMemberIDs, gotIDs)

			if tc.wantReorderArg != nil {
				require.Len(t, tc.memberRepo.reorderCalls, 1)
				assert.Equal(t, tc.wantReorderArg, tc.memberRepo.reorderCalls[0])
			}
		})
	}
}
