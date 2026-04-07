package application

import (
	"context"

	"github.com/dakotalillie/rota/internal/domain"
)

type DeleteOverrideInput struct {
	RotationID string
	OverrideID string
}

type DeleteOverrideUseCase struct {
	transactor   Transactor
	rotationRepo domain.RotationRepository
	overrideRepo domain.OverrideRepository
}

func NewDeleteOverrideUseCase(
	transactor Transactor,
	rotationRepo domain.RotationRepository,
	overrideRepo domain.OverrideRepository,
) *DeleteOverrideUseCase {
	return &DeleteOverrideUseCase{
		transactor:   transactor,
		rotationRepo: rotationRepo,
		overrideRepo: overrideRepo,
	}
}

func (uc *DeleteOverrideUseCase) Execute(ctx context.Context, input DeleteOverrideInput) error {
	return uc.transactor.RunInTx(ctx, func(ctx context.Context) error {
		if _, err := uc.rotationRepo.GetByID(ctx, input.RotationID); err != nil {
			return err
		}

		return uc.overrideRepo.Delete(ctx, input.RotationID, input.OverrideID)
	})
}
