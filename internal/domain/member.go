package domain

import "time"

var MemberColors = [...]string{"violet", "sky", "emerald", "orange", "rose", "teal", "amber", "pink"}

type Member struct {
	ID              string
	RotationID      string
	User            User
	Position        int
	Color           string
	BecameCurrentAt time.Time
}
