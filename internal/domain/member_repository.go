package domain

import "context"

type MemberRepository interface {
	CountByRotationID(ctx context.Context, rotationID string) (int, error)
	Create(ctx context.Context, rotationID, userID string, order int) (*Member, error)
}
