package application

import (
	"context"
	"time"

	"github.com/dakotalillie/rota/internal/domain"
)

type CreateOverrideInput struct {
	RotationID string
	MemberID   string
	Start      time.Time
	End        time.Time
}

type CreateOverrideUseCase struct {
	transactor   Transactor
	rotationRepo domain.RotationRepository
	overrideRepo domain.OverrideRepository
}

func NewCreateOverrideUseCase(
	transactor Transactor,
	rotationRepo domain.RotationRepository,
	overrideRepo domain.OverrideRepository,
) *CreateOverrideUseCase {
	return &CreateOverrideUseCase{
		transactor:   transactor,
		rotationRepo: rotationRepo,
		overrideRepo: overrideRepo,
	}
}

func (uc *CreateOverrideUseCase) Execute(ctx context.Context, input CreateOverrideInput) (*domain.Override, error) {
	var override *domain.Override
	err := uc.transactor.RunInTx(ctx, func(ctx context.Context) error {
		rotation, err := uc.rotationRepo.GetByID(ctx, input.RotationID)
		if err != nil {
			return err
		}

		if err := rotation.ValidateOverride(input.MemberID, input.Start, input.End); err != nil {
			return err
		}

		overlapping, err := uc.overrideRepo.HasOverlapping(ctx, input.RotationID, input.Start, input.End)
		if err != nil {
			return err
		}
		if overlapping {
			return domain.ErrOverrideConflict
		}

		override, err = uc.overrideRepo.Create(ctx, input.RotationID, input.MemberID, input.Start, input.End)
		if err != nil {
			return err
		}

		// Hydrate member (with user) from the rotation already loaded above.
		for _, m := range rotation.Members {
			if m.ID == input.MemberID {
				override.Member = m
				break
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return override, nil
}
