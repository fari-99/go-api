package keycloaks

type ListUser struct {
	ErrorResponse

	Id                         string             `json:"id"`
	CreatedTimestamp           int64              `json:"createdTimestamp"`
	Username                   string             `json:"username"`
	Enabled                    bool               `json:"enabled"`
	Totp                       bool               `json:"totp"`
	EmailVerified              bool               `json:"emailVerified"`
	FirstName                  string             `json:"firstName"`
	LastName                   string             `json:"lastName"`
	Email                      string             `json:"email"`
	Attributes                 ListUserAttributes `json:"attributes"`
	DisableableCredentialTypes []interface{}      `json:"disableableCredentialTypes"`
	RequiredActions            []string           `json:"requiredActions"`
	NotBefore                  int                `json:"notBefore"`
	Access                     ListUserAccess     `json:"access"`
}

type ListUserAttributes struct {
	PhoneNumberVerified    []string `json:"phoneNumberVerified"`
	SecondaryEmail         []string `json:"secondaryEmail"`
	ProfilePictures        []string `json:"profilePictures"`
	PhoneNumbers           []string `json:"phoneNumbers"`
	SecondaryEmailVerified []string `json:"secondaryEmailVerified"`
}

type ListUserAccess struct {
	ErrorResponse

	ManageGroupMembership bool `json:"manageGroupMembership"`
	View                  bool `json:"view"`
	MapRoles              bool `json:"mapRoles"`
	Impersonate           bool `json:"impersonate"`
	Manage                bool `json:"manage"`
}

type FilterListUser struct {
	BriefRepresentation string `json:"briefRepresentation"`
	Email               string `json:"email"`
	EmailVerified       string `json:"emailVerified"`
	Enabled             string `json:"enabled"`
	Exact               string `json:"exact"`
	First               string `json:"first"`
	FirstName           string `json:"firstName"`
	IdpAlias            string `json:"idpAlias"`
	IdpUserId           string `json:"idpUserId"`
	LastName            string `json:"lastName"`
	Max                 string `json:"max"`
	Q                   string `json:"q"`
	Search              string `json:"search"`
	Username            string `json:"username"`
}
