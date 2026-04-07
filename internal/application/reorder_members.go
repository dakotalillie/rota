package application

import (
	"context"
	"sort"

	"github.com/dakotalillie/rota/internal/domain"
)

type ReorderMembersInput struct {
	RotationID string
	MemberIDs  []string
}

type ReorderMembersUseCase struct {
	transactor   Transactor
	rotationRepo domain.RotationRepository
	memberRepo   domain.MemberRepository
}

func NewReorderMembersUseCase(
	transactor Transactor,
	rotationRepo domain.RotationRepository,
	memberRepo domain.MemberRepository,
) *ReorderMembersUseCase {
	return &ReorderMembersUseCase{
		transactor:   transactor,
		rotationRepo: rotationRepo,
		memberRepo:   memberRepo,
	}
}

func (uc *ReorderMembersUseCase) Execute(ctx context.Context, input ReorderMembersInput) (*domain.Rotation, error) {
	var rotation *domain.Rotation
	err := uc.transactor.RunInTx(ctx, func(ctx context.Context) error {
		var err error
		rotation, err = uc.rotationRepo.GetByID(ctx, input.RotationID)
		if err != nil {
			return err
		}

		if err := validateMemberIDs(rotation.Members, input.MemberIDs); err != nil {
			return err
		}

		return uc.memberRepo.ReorderMembers(ctx, input.RotationID, input.MemberIDs)
	})
	if err != nil {
		return nil, err
	}

	// Reload to pick up the updated order values.
	rotation, err = uc.rotationRepo.GetByID(ctx, input.RotationID)
	if err != nil {
		return nil, err
	}
	sort.Slice(rotation.Members, func(i, j int) bool {
		return rotation.Members[i].Order < rotation.Members[j].Order
	})
	return rotation, nil
}

// validateMemberIDs checks that the provided IDs are exactly the set of
// current member IDs for the rotation — same count, no duplicates, no unknowns.
func validateMemberIDs(members []domain.Member, ids []string) error {
	if len(ids) != len(members) {
		return domain.ErrMemberMismatch
	}
	existing := make(map[string]struct{}, len(members))
	for _, m := range members {
		existing[m.ID] = struct{}{}
	}
	seen := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		if _, ok := existing[id]; !ok {
			return domain.ErrMemberMismatch
		}
		if _, dup := seen[id]; dup {
			return domain.ErrMemberMismatch
		}
		seen[id] = struct{}{}
	}
	return nil
}
