package keycloaks

type AccountUserProfile struct {
	Id            string                       `json:"id"`
	Username      string                       `json:"username"`
	FirstName     string                       `json:"firstName"`
	LastName      string                       `json:"lastName"`
	Email         string                       `json:"email"`
	EmailVerified bool                         `json:"emailVerified"`
	Attributes    AccountUserProfileAttributes `json:"attributes"`

	ErrorResponse
}

type AccountUserProfileAttributes struct {
	SecondaryEmail         []string `json:"secondaryEmail"`
	SecondaryEmailVerified []string `json:"secondaryEmailVerified,omitempty"`
	PhoneNumbers           []string `json:"phoneNumbers"`
	PhoneNumberVerified    []string `json:"phoneNumberVerified,omitempty"`
	ProfilePictures        []string `json:"profilePictures"`
	ClientAccess           []string `json:"clientAccess"`
}
