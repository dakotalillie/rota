package presentation

type Rotation struct {
	Type       string             `json:"type"`
	ID         string             `json:"id"`
	Attributes RotationAttributes `json:"attributes"`
}

type RotationAttributes struct {
	Name    string          `json:"name"`
	Cadence RotationCadence `json:"cadence"`
}

type RotationCadence struct {
	Weekly *RotationCadenceWeekly `json:"weekly,omitempty"`
}

type RotationCadenceWeekly struct {
	Day      string `json:"day"`
	Time     string `json:"time"`
	TimeZone string `json:"timeZone"`
}
