package application

import (
	"context"

	"github.com/dakotalillie/rota/internal/domain"
)

type GetRotationUseCase struct {
	repo         domain.RotationRepository
	overrideRepo domain.OverrideRepository
	clock        domain.Clock
}

func NewGetRotationUseCase(repo domain.RotationRepository, overrideRepo domain.OverrideRepository, clock domain.Clock) *GetRotationUseCase {
	return &GetRotationUseCase{repo: repo, overrideRepo: overrideRepo, clock: clock}
}

func (uc *GetRotationUseCase) Execute(ctx context.Context, id string) (*domain.Rotation, error) {
	rotation, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	overrides, err := uc.overrideRepo.ListByRotationID(ctx, id, uc.clock.Now())
	if err != nil {
		return nil, err
	}
	rotation.Overrides = overrides

	return rotation, nil
}
