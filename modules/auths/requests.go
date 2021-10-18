package auths

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type RequestAuthUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (request RequestAuthUser) Validate() error {
	return validation.ValidateStruct(&request,
		validation.Field(&request.Email, validation.Required, is.Email),
		validation.Field(&request.Password, validation.Required))
}
