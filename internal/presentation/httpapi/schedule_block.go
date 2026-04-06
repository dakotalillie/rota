package httpapi

type ScheduleBlock struct {
	Type          string                     `json:"type"`
	ID            string                     `json:"id"`
	Attributes    ScheduleBlockAttributes    `json:"attributes"`
	Relationships ScheduleBlockRelationships `json:"relationships"`
}

type ScheduleBlockAttributes struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type ScheduleBlockRelationships struct {
	Member ScheduleBlockMemberRelationship `json:"member"`
}

type ScheduleBlockMemberRelationship struct {
	Data RelationshipData `json:"data"`
}
