package roles

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"go-api/constant"
)

type RequestListFilter struct {
	Page    int    `json:"page"`
	Limit   int    `json:"limit"`
	OrderBy string `json:"order_by"`
}

type RequestRole struct {
	RoleType int8   `json:"role_type"`
	RoleName string `json:"role_name"`
}

func (request RequestRole) Validate() error {
	return validation.ValidateStruct(&request,
		validation.Field(&request.RoleName, validation.Required),
		validation.Field(&request.RoleType, validation.Required,
			validation.In(int8(constant.UserTypeCustomer), int8(constant.UserTypeSeller))),
	)
}
