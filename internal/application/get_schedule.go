package application

import (
	"context"
	"time"

	"github.com/dakotalillie/rota/internal/domain"
)

type GetScheduleUseCase struct {
	repo         domain.RotationRepository
	overrideRepo domain.OverrideRepository
}

func NewGetScheduleUseCase(repo domain.RotationRepository, overrideRepo domain.OverrideRepository) *GetScheduleUseCase {
	return &GetScheduleUseCase{repo: repo, overrideRepo: overrideRepo}
}

func (uc *GetScheduleUseCase) Execute(ctx context.Context, rotationID string, now time.Time, numWeeks int) ([]domain.ScheduleBlock, error) {
	rotation, err := uc.repo.GetByID(ctx, rotationID)
	if err != nil {
		return nil, err
	}

	overrides, err := uc.overrideRepo.ListByRotationID(ctx, rotationID, now)
	if err != nil {
		return nil, err
	}
	rotation.Overrides = overrides

	return rotation.Schedule(now, numWeeks)
}
