package application

import (
	"context"
	"time"

	"github.com/dakotalillie/rota/internal/domain"
)

type ListRotationsUseCase struct {
	repo         domain.RotationRepository
	overrideRepo domain.OverrideRepository
}

func NewListRotationsUseCase(repo domain.RotationRepository, overrideRepo domain.OverrideRepository) *ListRotationsUseCase {
	return &ListRotationsUseCase{repo: repo, overrideRepo: overrideRepo}
}

func (uc *ListRotationsUseCase) Execute(ctx context.Context) ([]*domain.Rotation, error) {
	rotations, err := uc.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
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
