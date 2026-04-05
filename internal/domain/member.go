package domain

type Member struct {
	ID         string
	RotationID string
	User       User
	Order      int
}
