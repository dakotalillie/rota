package domain

import "context"

type RotationRepository interface {
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, rot *Rotation) (*Rotation, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*Rotation, error)
	List(ctx context.Context) ([]*Rotation, error)
}
