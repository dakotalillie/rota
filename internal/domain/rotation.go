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
	Day      string
	Time     string
	TimeZone string
}
