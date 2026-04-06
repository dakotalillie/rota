package application

import (
	"context"
	"time"

	"github.com/dakotalillie/rota/internal/domain"
)

type GetScheduleUseCase struct {
	repo domain.RotationRepository
}

func NewGetScheduleUseCase(repo domain.RotationRepository) *GetScheduleUseCase {
	return &GetScheduleUseCase{repo: repo}
}

func (uc *GetScheduleUseCase) Execute(ctx context.Context, rotationID string, now time.Time, numWeeks int) ([]domain.ScheduleBlock, error) {
	rotation, err := uc.repo.GetByID(ctx, rotationID)
	if err != nil {
		return nil, err
	}
	return rotation.Schedule(now, numWeeks)
}
