package domain

import (
	"context"
	"time"
)

type OverrideRepository interface {
	Create(ctx context.Context, rotationID, memberID string, start, end time.Time) (*Override, error)
	Delete(ctx context.Context, rotationID, overrideID string) error
	DeleteByMemberID(ctx context.Context, memberID string) error
	HasOverlapping(ctx context.Context, rotationID string, start, end time.Time) (bool, error)
	ListByRotationID(ctx context.Context, rotationID string, now time.Time) ([]Override, error)
	ListByRotationIDs(ctx context.Context, rotationIDs []string, now time.Time) (map[string][]Override, error)
}
