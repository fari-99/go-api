package keycloaks

import validation "github.com/go-ozzo/ozzo-validation/v4"

type InputGrantPassword struct {
	NewPassword string `json:"new_password"`
}

func (model InputGrantPassword) Validate() error {
	return validation.ValidateStruct(&model,
		validation.Field(&model.NewPassword, validation.Required),
	)
}

type InputGrantChangePhoneNumber struct {
	NewPhoneNumber string `json:"new_phone_number"`
}

func (model InputGrantChangePhoneNumber) Validate() error {
	return validation.ValidateStruct(&model,
		validation.Field(&model.NewPhoneNumber, validation.Required),
	)
}

type AuthenticateKeycloak struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret,omitempty"`
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	GrantType    string `json:"grant_type,omitempty"`
	Scope        string `json:"scope,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Code         string `json:"code,omitempty"`
	CodeVerifier string `json:"code_verifier"`
	RedirectUri  string `json:"redirect_uri,omitempty"`
}

type AuthenticateResponse struct {
	// If Failed
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`

	// If Success
	AccessToken      string `json:"access_token,omitempty"`
	ExpiresIn        int    `json:"expires_in,omitempty"`
	IdToken          string `json:"id_token,omitempty"`
	NotBeforePolicy  int    `json:"not-before-policy,omitempty"`
	RefreshExpiresIn int    `json:"refresh_expires_in,omitempty"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	Scope            string `json:"scope,omitempty"`
	SessionState     string `json:"session_state,omitempty"`
	TokenType        string `json:"token_type,omitempty"`
}

func (model AuthenticateKeycloak) ValidateGrantPassword() error {
	return validation.ValidateStruct(&model,
		validation.Field(&model.ClientID, validation.Required),
		validation.Field(&model.ClientSecret, validation.Required),
		validation.Field(&model.Username, validation.Required),
		validation.Field(&model.Password, validation.Required),
		validation.Field(&model.GrantType, validation.Required),
		validation.Field(&model.Scope, validation.Required),
	)
}

func (model AuthenticateKeycloak) ValidateGrantAuthoritizationCode() error {
	return validation.ValidateStruct(&model,
		validation.Field(&model.ClientID, validation.Required),
		validation.Field(&model.GrantType, validation.Required),
		validation.Field(&model.Scope, validation.Required),
		validation.Field(&model.Code, validation.Required),
		validation.Field(&model.CodeVerifier, validation.Required),
		validation.Field(&model.RedirectUri, validation.Required),
	)
}

type LogoutKeycloak struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret,omitempty"`
	RefreshToken string `json:"refresh_token"`
}

func (model LogoutKeycloak) Validate() error {
	return validation.ValidateStruct(&model,
		validation.Field(&model.ClientID, validation.Required),
		validation.Field(&model.RefreshToken, validation.Required),
	)
}

type LogoutKeycloakResponse struct {
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
}
