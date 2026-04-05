package application

import (
	"context"

	"github.com/dakotalillie/rota/internal/domain"
)

type GetRotationUseCase struct {
	repo domain.RotationRepository
}

func NewGetRotationUseCase(repo domain.RotationRepository) *GetRotationUseCase {
	return &GetRotationUseCase{repo: repo}
}

func (uc *GetRotationUseCase) Execute(ctx context.Context, id string) (*domain.Rotation, error) {
	return uc.repo.GetByID(ctx, id)
}
