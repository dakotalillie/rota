package domain

type Rotation struct {
	ID      string
	Name    string
	Cadence RotationCadence
}

type RotationCadence struct {
	Weekly *RotationCadenceWeekly
}

type RotationCadenceWeekly struct {
	RotationDay      string
	RotationTime     string
	RotationTimeZone string
}
