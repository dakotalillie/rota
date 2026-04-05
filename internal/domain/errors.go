package domain

import "errors"

var (
	ErrRotationNotFound       = errors.New("rotation not found")
	ErrUserNotFound           = errors.New("user not found")
	ErrMemberAlreadyExists    = errors.New("user is already a member of this rotation")
	ErrRotationMembershipFull = errors.New("rotation has reached the maximum number of members")
	ErrMissingUserFields      = errors.New("name and email are required when not specifying a user ID")
)
