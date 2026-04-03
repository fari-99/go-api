package keycloaks

type UserDetails struct {
	ErrorResponse

	Id                         string         `json:"id"`
	CreatedTimestamp           int64          `json:"createdTimestamp"`
	Username                   string         `json:"username"`
	Enabled                    bool           `json:"enabled"`
	Totp                       bool           `json:"totp"`
	EmailVerified              bool           `json:"emailVerified"`
	FirstName                  string         `json:"firstName"`
	LastName                   string         `json:"lastName"`
	Email                      string         `json:"email"`
	Attributes                 UserAttributes `json:"attributes"`
	DisableableCredentialTypes []interface{}  `json:"disableableCredentialTypes"`
	RequiredActions            []interface{}  `json:"requiredActions"`
	FederatedIdentities        []interface{}  `json:"federatedIdentities"`
	NotBefore                  int            `json:"notBefore"`
	Access                     UserAccess     `json:"access"`
	UserProfileMetadata        UserMetaData   `json:"userProfileMetadata"`
}

type UserAttributes struct {
	ClientAccess           []string `json:"clientAccess"`
	SecondaryEmail         []string `json:"secondaryEmail"`
	ProfilePictures        []string `json:"profilePictures"`
	PhoneNumberVerified    []string `json:"phoneNumberVerified"`
	PhoneNumbers           []string `json:"phoneNumbers"`
	PhoneNumber2Fa         []string `json:"phoneNumber2FA"`
	SecondaryEmailVerified []string `json:"secondaryEmailVerified"`
}

type UserAccess struct {
	ManageGroupMembership bool `json:"manageGroupMembership"`
	View                  bool `json:"view"`
	MapRoles              bool `json:"mapRoles"`
	Impersonate           bool `json:"impersonate"`
	Manage                bool `json:"manage"`
}

type UserGroups struct {
	Name               string `json:"name"`
	DisplayHeader      string `json:"displayHeader"`
	DisplayDescription string `json:"displayDescription"`
	Annotations        struct {
	} `json:"annotations"`
}

type UserMetaDataAttributes struct {
	Name        string                 `json:"name"`
	DisplayName string                 `json:"displayName"`
	Required    bool                   `json:"required"`
	ReadOnly    bool                   `json:"readOnly"`
	Validators  UserMetaDataValidators `json:"validators"`
	Annotations struct {
		InputType string `json:"inputType,omitempty"`
	} `json:"annotations,omitempty"`
	Group string `json:"group,omitempty"`
}

type UserMetaData struct {
	Attributes []UserMetaDataAttributes `json:"attributes"`
	Groups     []UserGroups             `json:"groups"`
}

type UserMetaDataValidators struct {
	UsernameProhibitedCharacters struct {
		IgnoreEmptyValue bool `json:"ignore.empty.value"`
	} `json:"username-prohibited-characters,omitempty"`
	Length struct {
		Max              int  `json:"max"`
		IgnoreEmptyValue bool `json:"ignore.empty.value"`
		Min              int  `json:"min,omitempty"`
	} `json:"length,omitempty"`
	Email struct {
		IgnoreEmptyValue bool   `json:"ignore.empty.value"`
		MaxLocalLength   string `json:"max-local-length,omitempty"`
	} `json:"email,omitempty"`
	PersonNameProhibitedCharacters struct {
		IgnoreEmptyValue bool `json:"ignore.empty.value"`
	} `json:"person-name-prohibited-characters,omitempty"`
	Pattern struct {
		Pattern          string `json:"pattern"`
		ErrorMessage     string `json:"error-message"`
		IgnoreEmptyValue bool   `json:"ignore.empty.value"`
	} `json:"pattern,omitempty"`
}

type VerifyEmail struct {
	ClientID    string `json:"client_id,omitempty"`
	RedirectUri string `json:"redirect_uri,omitempty"`
}
