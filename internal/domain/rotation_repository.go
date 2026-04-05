package domain

import "context"

type RotationRepository interface {
	GetRotationByID(ctx context.Context, id string) (*Rotation, error)
}
