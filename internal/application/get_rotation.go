package application

import "github.com/dakotalillie/rota/internal/domain"

type GetRotationUseCase struct {
}

func (uc *GetRotationUseCase) Execute(id string) (*domain.Rotation, error) {
	return nil, nil
}

func NewGetRotationUseCase() *GetRotationUseCase {
	return &GetRotationUseCase{}
}
