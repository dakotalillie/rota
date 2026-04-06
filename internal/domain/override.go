package domain

import "time"

type Override struct {
	ID         string
	RotationID string
	Member     Member
	Start      time.Time
	End        time.Time
}
