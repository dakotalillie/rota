package application

import (
	"context"
	"sort"
	"time"

	"github.com/dakotalillie/rota/internal/domain"
)

type DeleteMemberInput struct {
	RotationID string
	MemberID   string
	Now        time.Time // used to set became_current_at when promoting the next member
}

type DeleteMemberUseCase struct {
	transactor   Transactor
	rotationRepo domain.RotationRepository
	memberRepo   domain.MemberRepository
	userRepo     domain.UserRepository
}

func NewDeleteMemberUseCase(
	transactor Transactor,
	rotationRepo domain.RotationRepository,
	memberRepo domain.MemberRepository,
	userRepo domain.UserRepository,
) *DeleteMemberUseCase {
	return &DeleteMemberUseCase{
		transactor:   transactor,
		rotationRepo: rotationRepo,
		memberRepo:   memberRepo,
		userRepo:     userRepo,
	}
}

func (uc *DeleteMemberUseCase) Execute(ctx context.Context, input DeleteMemberInput) error {
	return uc.transactor.RunInTx(ctx, func(ctx context.Context) error {
		rotation, err := uc.rotationRepo.GetByID(ctx, input.RotationID)
		if err != nil {
			return err
		}

		member, err := uc.memberRepo.GetByID(ctx, input.RotationID, input.MemberID)
		if err != nil {
			return err
		}

		if err := uc.memberRepo.Delete(ctx, input.MemberID); err != nil {
			return err
		}

		count, err := uc.userRepo.CountMemberships(ctx, member.User.ID)
		if err != nil {
			return err
		}
		if count == 0 {
			if err := uc.userRepo.Delete(ctx, member.User.ID); err != nil {
				return err
			}
		}

		// Sort the full member list by order so we can compute positions.
		sorted := make([]domain.Member, len(rotation.Members))
		copy(sorted, rotation.Members)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Order < sorted[j].Order
		})

		// Build remaining IDs in order (excluding the deleted member).
		remainingIDs := make([]string, 0, len(sorted)-1)
		deletedIndex := -1
		for i, m := range sorted {
			if m.ID == input.MemberID {
				deletedIndex = i
				continue
			}
			remainingIDs = append(remainingIDs, m.ID)
		}

		if len(remainingIDs) > 0 {
			if err := uc.memberRepo.ReorderMembers(ctx, input.RotationID, remainingIDs); err != nil {
				return err
			}
		}

		// If the deleted member was current, promote the next one in order (wrapping).
		if rotation.CurrentMember != nil && rotation.CurrentMember.ID == input.MemberID && len(remainingIDs) > 0 {
			nextIndex := deletedIndex % len(remainingIDs)
			nextMemberID := remainingIDs[nextIndex]
			if err := uc.memberRepo.SetCurrentMember(ctx, input.RotationID, nextMemberID, input.Now); err != nil {
				return err
			}
		}

		return nil
	})
}
