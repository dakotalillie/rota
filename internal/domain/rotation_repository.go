package domain

type RotationRepository interface {
	GetRotationByID(id string) (*Rotation, error)
}
