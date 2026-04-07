package domain

import "context"

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	Create(ctx context.Context, name, email string) (*User, error)
	CountMemberships(ctx context.Context, userID string) (int, error)
	Delete(ctx context.Context, userID string) error
}
