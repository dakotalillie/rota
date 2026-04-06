package domain

import "errors"

var (
	ErrRotationNotFound       = errors.New("rotation not found")
	ErrTooManyRotations       = errors.New("maximum number of rotations reached")
	ErrInvalidRotationName    = errors.New("rotation name is required")
	ErrUserNotFound           = errors.New("user not found")
	ErrMemberAlreadyExists    = errors.New("user is already a member of this rotation")
	ErrRotationMembershipFull = errors.New("rotation has reached the maximum number of members")
	ErrMissingUserFields      = errors.New("name and email are required when not specifying a user ID")
	ErrMemberNotFound         = errors.New("member not found in rotation")
	ErrOverrideSameMember     = errors.New("override member is already the scheduled on-call during this window")
	ErrOverrideConflict       = errors.New("an override already exists during this window")
)
