package httpapi

type Rotation struct {
	Type          string                `json:"type"`
	ID            string                `json:"id"`
	Links         RotationLinks         `json:"links"`
	Attributes    RotationAttributes    `json:"attributes"`
	Relationships RotationRelationships `json:"relationships"`
}

type RotationLinks struct {
	Self     string `json:"self"`
	Schedule string `json:"schedule"`
}

type RotationRelationships struct {
	CurrentMember   CurrentMemberRelationship   `json:"currentMember"`
	ScheduledMember ScheduledMemberRelationship `json:"scheduledMember"`
	Members         *MembersRelationship        `json:"members,omitempty"`
	Overrides       *OverridesRelationship      `json:"overrides,omitempty"`
}

type OverridesRelationship struct {
	Data []RelationshipData `json:"data"`
}

type MembersRelationship struct {
	Data []RelationshipData `json:"data"`
}

type CurrentMemberRelationship struct {
	Data *RelationshipData `json:"data"`
}

type ScheduledMemberRelationship struct {
	Data *RelationshipData `json:"data"`
}

type RelationshipData struct {
	Type string `json:"type"`
	ID   string `json:"id"`
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
