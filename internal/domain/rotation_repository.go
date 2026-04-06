package domain

import "context"

type RotationRepository interface {
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, rot *Rotation) (*Rotation, error)
	GetByID(ctx context.Context, id string) (*Rotation, error)
}
