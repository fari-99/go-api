package permissions

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Policy struct {
	PType   string `json:"p_type"`
	Subject string `json:"subject"`
	Route   string `json:"route"`
	Method  string `json:"method"`
}

type CreatePermissions Policy

func (m CreatePermissions) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.PType, validation.Required, validation.In("p", "g")),
		validation.Field(&m.Subject, validation.Required),
		validation.Field(&m.Route, validation.Required),
		validation.Field(&m.Method, validation.Required),
	)
}

type DeletePermissions Policy

func (m DeletePermissions) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.PType, validation.Required, validation.In("p", "g")),
		validation.Field(&m.Subject, validation.Required),
		validation.Field(&m.Route, validation.Required),
		validation.Field(&m.Method, validation.Required),
	)
}

type EditPermissions struct {
	NewPolicy Policy `json:"new_policy"`
	OldPolicy Policy `json:"old_policy"`
}

func (m EditPermissions) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.OldPolicy, validation.Required),
		validation.Field(&m.NewPolicy, validation.Required),
		validation.Field(&m.OldPolicy.PType, validation.Required, validation.In("p", "g")),
		validation.Field(&m.OldPolicy.Subject, validation.Required),
		validation.Field(&m.OldPolicy.Route, validation.Required),
		validation.Field(&m.OldPolicy.Method, validation.Required),
		validation.Field(&m.NewPolicy.PType, validation.Required, validation.In("p", "g")),
		validation.Field(&m.NewPolicy.Subject, validation.Required),
		validation.Field(&m.NewPolicy.Route, validation.Required),
		validation.Field(&m.NewPolicy.Method, validation.Required),
	)
}

type CheckPermissions struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

func (m CheckPermissions) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Path, validation.Required),
		validation.Field(&m.Method, validation.Required),
	)
}
