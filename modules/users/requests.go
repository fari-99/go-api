package users

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type RequestCreateUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func (request RequestCreateUser) Validate() error {
	return validation.ValidateStruct(&request,
		validation.Field(&request.Email, validation.Required, is.Email),
		validation.Field(&request.Username, validation.Required),
		validation.Field(&request.Password, validation.Required))
}

type RequestChangePassword struct {
	CurrentPassword    string `json:"current_password"`
	NewPassword        string `json:"new_password"`
	NewPasswordConfirm string `json:"new_password_confirm"`
}

func (request RequestChangePassword) Validate() error {
	return validation.ValidateStruct(&request,
		validation.Field(&request.CurrentPassword, validation.Required),
		validation.Field(&request.NewPassword, validation.Required),
		validation.Field(&request.NewPasswordConfirm, validation.Required),
	)
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

func (request ForgotPasswordRequest) Validate() error {
	return validation.ValidateStruct(&request,
		validation.Field(&request.Email, validation.Required),
	)
}

type ForgotUsernameRequest struct {
	Email string `json:"email"`
}

func (request ForgotUsernameRequest) Validate() error {
	return validation.ValidateStruct(&request,
		validation.Field(&request.Email, validation.Required),
	)
}

type ResetPasswordRequest struct {
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
	Token                string `json:"token"`

	HashedPassword string `json:"-"`
}

func (request ResetPasswordRequest) Validate() error {
	err := validation.ValidateStruct(&request,
		validation.Field(&request.Password, validation.Required, validation.Length(8, 255)),
		validation.Field(&request.PasswordConfirmation, validation.Required, validation.In(request.Password).Error("Your 'password' and 'password_confirmation' do not match")),
		validation.Field(&request.Token, validation.Required),
	)
	if err != nil && request.Password != request.PasswordConfirmation {
		return err
	}

	return nil
}
