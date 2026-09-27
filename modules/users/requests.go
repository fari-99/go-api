package users

import (
	"errors"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"

	"go-api/constant"
)

type RequestListUsers struct {
	Page    int    `json:"page"`
	Limit   int    `json:"limit"`
	OrderBy string `json:"order_by"`
	Search  string `json:"search"`
}

type RequestUpdateUser struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	MobilePhone string `json:"mobile_phone"`
	Status      int8   `json:"status"`
}

func (request RequestUpdateUser) Validate() error {
	return validation.ValidateStruct(&request,
		validation.Field(&request.Email, validation.Required, is.Email),
		validation.Field(&request.Username, validation.Required),
		validation.Field(&request.Status, validation.Required,
			validation.In(int8(constant.StatusActive), int8(constant.StatusNonActive), int8(constant.StatusDeleted))),
	)
}

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
		validation.Field(&request.NewPasswordConfirm, validation.Required, validation.By(func(value interface{}) error {
			confirm, _ := value.(string)
			if confirm != request.NewPassword {
				return errors.New("new_password_confirm does not match new_password")
			}
			return nil
		})),
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

type RequestUserRoles struct {
	RoleIDs []uint64 `json:"role_ids"`
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
