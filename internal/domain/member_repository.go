package domain

import (
	"context"
	"time"
)

type MemberRepository interface {
	CountByRotationID(ctx context.Context, rotationID string) (int, error)
	Create(ctx context.Context, rotationID, userID string, order int) (*Member, error)
	GetByID(ctx context.Context, rotationID, memberID string) (*Member, error)
	Delete(ctx context.Context, memberID string) error
	SetCurrentMember(ctx context.Context, rotationID string, memberID string, at time.Time) error
	ReorderMembers(ctx context.Context, rotationID string, memberIDs []string) error
}
