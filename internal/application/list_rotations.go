package application

import (
	"context"

	"github.com/dakotalillie/rota/internal/domain"
)

type ListRotationsUseCase struct {
	repo domain.RotationRepository
}

func NewListRotationsUseCase(repo domain.RotationRepository) *ListRotationsUseCase {
	return &ListRotationsUseCase{repo: repo}
}

func (uc *ListRotationsUseCase) Execute(ctx context.Context) ([]*domain.Rotation, error) {
	return uc.repo.List(ctx)
}
