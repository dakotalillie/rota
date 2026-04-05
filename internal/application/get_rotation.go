package application

import "github.com/dakotalillie/rota/internal/domain"

type GetRotationUseCase struct {
	repo domain.RotationRepository
}

func NewGetRotationUseCase(repo domain.RotationRepository) *GetRotationUseCase {
	return &GetRotationUseCase{repo: repo}
}

func (uc *GetRotationUseCase) Execute(id string) (*domain.Rotation, error) {
	return uc.repo.GetRotationByID(id)
}
