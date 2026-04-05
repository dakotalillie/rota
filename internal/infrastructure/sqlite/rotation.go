package sqlite

type rotationData struct {
	Name    string          `json:"name"`
	Cadence rotationCadence `json:"cadence"`
}

type rotationCadence struct {
	Weekly *rotationCadenceWeekly `json:"weekly,omitempty"`
}

type rotationCadenceWeekly struct {
	Day      string `json:"day"`
	Time     string `json:"time"`
	TimeZone string `json:"timeZone"`
}
