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

type fakeOverrideRepo struct {
	hasOverlapping        bool
	hasOverlappingErr     error
	createdOverride       *domain.Override
	createErr             error
	deleteByMemberIDCalls []string
	deleteByMemberIDErr   error
}

func (f *fakeOverrideRepo) Create(_ context.Context, rotationID, memberID string, start, end time.Time) (*domain.Override, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	if f.createdOverride != nil {
		return f.createdOverride, nil
	}
	return &domain.Override{
		ID:         "ovr_01JQGF0000000000000000001",
		RotationID: rotationID,
		Member:     domain.Member{ID: memberID},
		Start:      start,
		End:        end,
	}, nil
}

func (f *fakeOverrideRepo) HasOverlapping(_ context.Context, _ string, _, _ time.Time) (bool, error) {
	return f.hasOverlapping, f.hasOverlappingErr
}

func (f *fakeOverrideRepo) DeleteByMemberID(_ context.Context, memberID string) error {
	f.deleteByMemberIDCalls = append(f.deleteByMemberIDCalls, memberID)
	return f.deleteByMemberIDErr
}

func (f *fakeOverrideRepo) ListByRotationID(_ context.Context, _ string, _ time.Time) ([]domain.Override, error) {
	return []domain.Override{}, nil
}

func TestCreateOverrideUseCase_Execute(t *testing.T) {
	const rotationID = "rot_01JQGF0000000000000000000"
	const memberID = "mem_01JQGF0000000000000000001"

	aliceMember := domain.Member{
		ID:    memberID,
		Order: 1,
		User:  domain.User{ID: "usr_01JQGF0000000000000000000", Name: "Alice Smith", Email: "alice@example.com"},
	}
	bobMember := domain.Member{
		ID:    "mem_01JQGF0000000000000000002",
		Order: 2,
		User:  domain.User{ID: "usr_01JQGF0000000000000000001", Name: "Bob Jones", Email: "bob@example.com"},
	}

	// A rotation with two members but no active schedule (no current member)
	rotationNoSchedule := &domain.Rotation{
		ID:      rotationID,
		Name:    "Platform On-Call",
		Members: []domain.Member{aliceMember, bobMember},
	}

	start := time.Date(2026, 4, 7, 9, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 14, 9, 0, 0, 0, time.UTC)

	validInput := application.CreateOverrideInput{
		RotationID: rotationID,
		MemberID:   memberID,
		Start:      start,
		End:        end,
	}

	tests := []struct {
		name         string
		input        application.CreateOverrideInput
		rotationRepo fakeRotationRepo
		overrideRepo fakeOverrideRepo
		wantErr      error
		checkResult  func(t *testing.T, got *domain.Override)
	}{
		{
			name:         "success - member hydrated from rotation",
			input:        validInput,
			rotationRepo: fakeRotationRepo{rotation: rotationNoSchedule},
			overrideRepo: fakeOverrideRepo{},
			checkResult: func(t *testing.T, got *domain.Override) {
				require.NotNil(t, got)
				assert.Equal(t, rotationID, got.RotationID)
				assert.Equal(t, aliceMember, got.Member)
				assert.Equal(t, start, got.Start)
				assert.Equal(t, end, got.End)
			},
		},
		{
			name:         "rotation not found",
			input:        validInput,
			rotationRepo: fakeRotationRepo{err: domain.ErrRotationNotFound},
			wantErr:      domain.ErrRotationNotFound,
		},
		{
			name:  "member not in rotation",
			input: validInput,
			rotationRepo: fakeRotationRepo{rotation: &domain.Rotation{
				ID:      rotationID,
				Members: []domain.Member{bobMember}, // alice not present
			}},
			wantErr: domain.ErrMemberNotFound,
		},
		{
			name:         "override conflict",
			input:        validInput,
			rotationRepo: fakeRotationRepo{rotation: rotationNoSchedule},
			overrideRepo: fakeOverrideRepo{hasOverlapping: true},
			wantErr:      domain.ErrOverrideConflict,
		},
		{
			name:         "HasOverlapping error propagates",
			input:        validInput,
			rotationRepo: fakeRotationRepo{rotation: rotationNoSchedule},
			overrideRepo: fakeOverrideRepo{hasOverlappingErr: errors.New("db error")},
			wantErr:      errors.New("db error"),
		},
		{
			name:         "create error propagates",
			input:        validInput,
			rotationRepo: fakeRotationRepo{rotation: rotationNoSchedule},
			overrideRepo: fakeOverrideRepo{createErr: errors.New("db error")},
			wantErr:      errors.New("db error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uc := application.NewCreateOverrideUseCase(
				&fakeTransactor{},
				&tc.rotationRepo,
				&tc.overrideRepo,
			)

			got, err := uc.Execute(context.Background(), tc.input)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, tc.wantErr.Error(), err.Error())
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			if tc.checkResult != nil {
				tc.checkResult(t, got)
			}
		})
	}
}
