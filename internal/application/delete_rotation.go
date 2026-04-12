package application

import (
	"context"

	"github.com/dakotalillie/rota/internal/domain"
)

type DeleteRotationInput struct {
	RotationID string
}

type DeleteRotationUseCase struct {
	transactor   Transactor
	rotationRepo domain.RotationRepository
	memberRepo   domain.MemberRepository
	overrideRepo domain.OverrideRepository
	userRepo     domain.UserRepository
}

func NewDeleteRotationUseCase(
	transactor Transactor,
	rotationRepo domain.RotationRepository,
	memberRepo domain.MemberRepository,
	overrideRepo domain.OverrideRepository,
	userRepo domain.UserRepository,
) *DeleteRotationUseCase {
	return &DeleteRotationUseCase{
		transactor:   transactor,
		rotationRepo: rotationRepo,
		memberRepo:   memberRepo,
		overrideRepo: overrideRepo,
		userRepo:     userRepo,
	}
}

func (uc *DeleteRotationUseCase) Execute(ctx context.Context, input DeleteRotationInput) error {
	return uc.transactor.RunInTx(ctx, func(ctx context.Context) error {
		rotation, err := uc.rotationRepo.GetByID(ctx, input.RotationID)
		if err != nil {
			return err
		}

		for _, member := range rotation.Members {
			if err := uc.overrideRepo.DeleteByMemberID(ctx, member.ID); err != nil {
				return err
			}
			if err := uc.memberRepo.Delete(ctx, member.ID); err != nil {
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
		}

		return uc.rotationRepo.Delete(ctx, input.RotationID)
	})
}
