package domain

import "context"

type RotationRepository interface {
	GetByID(ctx context.Context, id string) (*Rotation, error)
}
