package application

import (
	"context"

	"github.com/dakotalillie/rota/internal/domain"
)

type ListRotationsUseCase struct {
	repo         domain.RotationRepository
	overrideRepo domain.OverrideRepository
	clock        domain.Clock
}

func NewListRotationsUseCase(repo domain.RotationRepository, overrideRepo domain.OverrideRepository, clock domain.Clock) *ListRotationsUseCase {
	return &ListRotationsUseCase{repo: repo, overrideRepo: overrideRepo, clock: clock}
}

func (uc *ListRotationsUseCase) Execute(ctx context.Context) ([]*domain.Rotation, error) {
	rotations, err := uc.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	now := uc.clock.Now()
	rotationIDs := make([]string, 0, len(rotations))
	for _, rotation := range rotations {
		rotationIDs = append(rotationIDs, rotation.ID)
	}

	overridesByRotation, err := uc.overrideRepo.ListByRotationIDs(ctx, rotationIDs, now)
	if err != nil {
		return nil, err
	}
	for _, rotation := range rotations {
		rotation.Overrides = overridesByRotation[rotation.ID]
	}

	return rotations, nil
}
