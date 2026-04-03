package keycloaks

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// InputCreateUser input create user to keycloak
/*
Example of data send to keycloak to create user
{
  "user_id":"1",
  "requiredActions": [],
  "emailVerified": false,
  "username": "testtest",
  "email": "testtest@gmail.com",
  "firstName": "testtest",
  "lastName": "testtest",
  "attributes": {
    "clientAccess": "",
    "secondaryEmail": "",
    "secondaryEmailVerified": "",
    "phoneNumbers": "",
    "phoneNumberVerified": "",
    "profilePictures": ""
  },
  "groups": [],
  "enabled": true
}
*/
type InputCreateUser struct {
	RequiredActions     []string                 `json:"requiredActions,omitempty"`
	EmailVerified       bool                     `json:"emailVerified,omitempty"`
	Username            string                   `json:"username,omitempty"`
	Email               string                   `json:"email,omitempty"`
	FirstName           string                   `json:"firstName,omitempty"`
	LastName            string                   `json:"lastName,omitempty"`
	Attributes          InputUserAttributes      `json:"attributes,omitempty"`
	Enabled             bool                     `json:"enabled,omitempty"`
	Credentials         []InputCredentials       `json:"credentials,omitempty"`
	FederatedIdentities []InputFederatedIdentity `json:"federatedIdentities,omitempty"`
}

func (model InputCreateUser) Validate() error {
	return validation.ValidateStruct(&model,
		validation.Field(&model.Username, validation.Required),
		validation.Field(&model.Email, validation.Required),
		validation.Field(&model.FirstName, validation.Required),
		validation.Field(&model.LastName, validation.Required),
	)
}

type InputUserAttributes struct {
	ClientAccess           []string `json:"clientAccess,omitempty"`
	SecondaryEmail         []string `json:"secondaryEmail,omitempty"`
	SecondaryEmailVerified []string `json:"secondaryEmailVerified,omitempty"`
	PhoneNumbers           []string `json:"phoneNumbers,omitempty"`
	PhoneNumber2Fa         []string `json:"phoneNumber2FA,omitempty"`
	PhoneNumberVerified    []string `json:"phoneNumberVerified,omitempty"`
	ProfilePictures        []string `json:"profilePictures,omitempty"`
	DmpUserId              []string `json:"dmpUserId,omitempty"`
}

type InputSetPassword struct {
	// Temporary
	// if false, then its permanent,
	// if true, then user will need to update their password (automatically enforced by keycloak when login)
	Temporary bool `json:"temporary"`

	// Type
	// Default value is "password"
	Type string `json:"type"`

	// Value of the password
	Value string `json:"value"`
}

func (model InputSetPassword) Validate() error {
	return validation.ValidateStruct(&model,
		validation.Field(&model.Type, validation.Required),
		validation.Field(&model.Value, validation.Required),
	)
}

type InputUpdateUser struct {
	RequiredActions []string            `json:"requiredActions,omitempty"`
	EmailVerified   bool                `json:"emailVerified,omitempty"`
	Username        string              `json:"username,omitempty"`
	Email           string              `json:"email,omitempty"`
	FirstName       string              `json:"firstName,omitempty"`
	LastName        string              `json:"lastName,omitempty"`
	Attributes      InputUserAttributes `json:"attributes,omitempty"`
	Enabled         bool                `json:"enabled,omitempty"`
}

func (model InputUpdateUser) Validate() error {
	return validation.ValidateStruct(&model,
		validation.Field(&model.Username, validation.Required),
		validation.Field(&model.Email, validation.Required),
		validation.Field(&model.FirstName, validation.Required),
		validation.Field(&model.LastName, validation.Required),
	)
}

type InputCredentials struct {
	ID                string `json:"id,omitempty"`
	Type              string `json:"type,omitempty"`
	UserLabel         string `json:"userLabel,omitempty"`
	CreatedDate       int    `json:"createdDate,omitempty"`
	SecretData        string `json:"secretData,omitempty"`
	CredentialData    string `json:"credentialData,omitempty"`
	Priority          int    `json:"priority,omitempty"`
	Value             string `json:"value,omitempty"`
	Temporary         bool   `json:"temporary,omitempty"`
	Device            string `json:"device,omitempty"`
	HashedSaltedValue string `json:"hashedSaltedValue,omitempty"`
	Salt              string `json:"salt,omitempty"`
	HashIterations    int    `json:"hashIterations,omitempty"`
	Counter           int    `json:"counter,omitempty"`
	Algorithm         string `json:"algorithm,omitempty"`
	Digits            int    `json:"digits,omitempty"`
	Period            int    `json:"period,omitempty"`
}

type InputFederatedIdentity struct {
	IdentityProvider string `json:"identityProvider"`
	UserId           string `json:"userId"`
	UserName         string `json:"userName"`
}

func (model InputFederatedIdentity) Validate() error {
	return validation.ValidateStruct(&model,
		validation.Field(&model.IdentityProvider, validation.Required),
		validation.Field(&model.UserId, validation.Required),
		validation.Field(&model.UserName, validation.Required),
	)
}
