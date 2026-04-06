package domain

import (
	"context"
	"time"
)

type OverrideRepository interface {
	Create(ctx context.Context, rotationID, memberID string, start, end time.Time) (*Override, error)
	HasOverlapping(ctx context.Context, rotationID string, start, end time.Time) (bool, error)
	ListByRotationID(ctx context.Context, rotationID string, now time.Time) ([]Override, error)
}
