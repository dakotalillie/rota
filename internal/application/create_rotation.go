package application

import (
	"context"

	"github.com/dakotalillie/rota/internal/domain"
)

type CreateRotationInput struct {
	Name string
}

type CreateRotationUseCase struct {
	transactor   Transactor
	rotationRepo domain.RotationRepository
}

func NewCreateRotationUseCase(transactor Transactor, rotationRepo domain.RotationRepository) *CreateRotationUseCase {
	return &CreateRotationUseCase{transactor: transactor, rotationRepo: rotationRepo}
}

func (uc *CreateRotationUseCase) Execute(ctx context.Context, input CreateRotationInput) (*domain.Rotation, error) {
	if input.Name == "" {
		return nil, domain.ErrInvalidRotationName
	}

	rot := &domain.Rotation{
		Name: input.Name,
		Cadence: domain.RotationCadence{
			Weekly: &domain.RotationCadenceWeekly{
				Day:      "monday",
				Time:     "09:00",
				TimeZone: "America/Los_Angeles",
			},
		},
	}

	var result *domain.Rotation
	err := uc.transactor.RunInTx(ctx, func(ctx context.Context) error {
		count, err := uc.rotationRepo.Count(ctx)
		if err != nil {
			return err
		}
		if count >= 20 {
			return domain.ErrTooManyRotations
		}

		result, err = uc.rotationRepo.Create(ctx, rot)
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
