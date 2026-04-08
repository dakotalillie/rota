package httpapi

type Member struct {
	Type          string              `json:"type"`
	ID            string              `json:"id"`
	Attributes    MemberAttributes    `json:"attributes"`
	Relationships MemberRelationships `json:"relationships"`
}

type MemberAttributes struct {
	Order int    `json:"order"`
	Color string `json:"color"`
}

type MemberRelationships struct {
	User MemberUserRelationship `json:"user"`
}

type MemberUserRelationship struct {
	Data MemberUserRelationshipData `json:"data"`
}

type MemberUserRelationshipData struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type IncludedUser struct {
	Type       string                 `json:"type"`
	ID         string                 `json:"id"`
	Attributes IncludedUserAttributes `json:"attributes"`
}

type IncludedUserAttributes struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
