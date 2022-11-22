package locations

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type RequestCreateLocations struct {
	ParentID string `json:"parent_id,omitempty"`
	Code     string `json:"code,omitempty"`
	Name     string `json:"name"`
	LevelID  string `json:"level_id"`
	Status   int8   `json:"status"`
}

func (request RequestCreateLocations) Validate() error {
	return validation.ValidateStruct(&request,
		validation.Field(&request.Name, validation.Required),
		validation.Field(&request.LevelID, validation.Required),
		validation.Field(&request.Status, validation.Required),
	)
}

type RequestUpdateLocations struct {
	ParentID string `json:"parent_id,omitempty"`
	Code     string `json:"code,omitempty"`
	Name     string `json:"name"`
	LevelID  string `json:"level_id"`
	Status   int8   `json:"status"`
}

func (request RequestUpdateLocations) Validate() error {
	return validation.ValidateStruct(&request,
		validation.Field(&request.Name, validation.Required),
		validation.Field(&request.LevelID, validation.Required),
		validation.Field(&request.Status, validation.Required),
	)
}

// ---------------------------------------------------
// LEVELS

type RequestCreateLocationLevel struct {
	NeedParentID *bool  `json:"need_parent_id"`
	Name         string `json:"name"`
	Status       int8   `json:"status"`
}

func (request RequestCreateLocationLevel) Validate() error {
	return validation.ValidateStruct(&request,
		validation.Field(&request.NeedParentID, validation.Nil),
		validation.Field(&request.Name, validation.Required),
		validation.Field(&request.Status, validation.Required),
	)
}

type RequestUpdateLocationLevel struct {
	NeedParentID *bool  `json:"need_parent_id"`
	Name         string `json:"name"`
	Status       int8   `json:"status"`
}

func (request RequestUpdateLocationLevel) Validate() error {
	return validation.ValidateStruct(&request,
		validation.Field(&request.NeedParentID, validation.Nil),
		validation.Field(&request.Name, validation.Required),
		validation.Field(&request.Status, validation.Required),
	)
}
