// Package application_test shared fakes used across use-case test files.
package application_test

import (
	"context"
	"time"

	"github.com/dakotalillie/rota/internal/domain"
)

type fakeTransactor struct{}

func (f *fakeTransactor) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type fakeRotationRepo struct {
	count       int
	countErr    error
	rotation    *domain.Rotation
	rotations   []*domain.Rotation
	err         error
	deleteErr   error
	deleteCalls []string
}

func (f *fakeRotationRepo) Count(_ context.Context) (int, error) {
	return f.count, f.countErr
}

func (f *fakeRotationRepo) Create(_ context.Context, rot *domain.Rotation) (*domain.Rotation, error) {
	return rot, f.err
}

func (f *fakeRotationRepo) Delete(_ context.Context, id string) error {
	f.deleteCalls = append(f.deleteCalls, id)
	return f.deleteErr
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
	count        int
	countErr     error
	createMember *domain.Member
	createErr    error
	createCalls  []struct {
		rotationID string
		userID     string
		position   int
		color      string
	}
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

func (f *fakeMemberRepo) Create(_ context.Context, rotationID, userID string, position int, color string) (*domain.Member, error) {
	f.createCalls = append(f.createCalls, struct {
		rotationID string
		userID     string
		position   int
		color      string
	}{rotationID, userID, position, color})
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

type fakeOverrideRepo struct {
	hasOverlapping      bool
	hasOverlappingErr   error
	createdOverride     *domain.Override
	createErr           error
	listByRotationID    map[string][]domain.Override
	listByRotationErr   error
	listByRotationCalls []struct {
		rotationID string
		now        time.Time
	}
	listByRotationIDsCalls []struct {
		rotationIDs []string
		now         time.Time
	}
	deleteCalls []struct {
		rotationID string
		overrideID string
	}
	deleteErr             error
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

func (f *fakeOverrideRepo) Delete(_ context.Context, rotationID, overrideID string) error {
	f.deleteCalls = append(f.deleteCalls, struct {
		rotationID string
		overrideID string
	}{rotationID, overrideID})
	return f.deleteErr
}

func (f *fakeOverrideRepo) DeleteByMemberID(_ context.Context, memberID string) error {
	f.deleteByMemberIDCalls = append(f.deleteByMemberIDCalls, memberID)
	return f.deleteByMemberIDErr
}

func (f *fakeOverrideRepo) ListByRotationID(_ context.Context, rotationID string, now time.Time) ([]domain.Override, error) {
	f.listByRotationCalls = append(f.listByRotationCalls, struct {
		rotationID string
		now        time.Time
	}{rotationID, now})
	if f.listByRotationErr != nil {
		return nil, f.listByRotationErr
	}
	if f.listByRotationID == nil {
		return []domain.Override{}, nil
	}
	return f.listByRotationID[rotationID], nil
}

func (f *fakeOverrideRepo) ListByRotationIDs(_ context.Context, rotationIDs []string, now time.Time) (map[string][]domain.Override, error) {
	f.listByRotationIDsCalls = append(f.listByRotationIDsCalls, struct {
		rotationIDs []string
		now         time.Time
	}{append([]string(nil), rotationIDs...), now})
	if f.listByRotationErr != nil {
		return nil, f.listByRotationErr
	}
	result := make(map[string][]domain.Override, len(rotationIDs))
	for _, rotationID := range rotationIDs {
		if f.listByRotationID == nil {
			result[rotationID] = []domain.Override{}
			continue
		}
		result[rotationID] = f.listByRotationID[rotationID]
		if result[rotationID] == nil {
			result[rotationID] = []domain.Override{}
		}
	}
	return result, nil
}
