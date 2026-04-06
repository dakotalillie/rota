package domain

import "time"

type Member struct {
	ID              string
	RotationID      string
	User            User
	Order           int
	BecameCurrentAt time.Time
}
