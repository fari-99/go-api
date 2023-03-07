package calendar_managements

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type RequestListFilter struct {
	Page    int    `json:"page"`
	Limit   int    `json:"limit"`
	OrderBy string `json:"order_by"`
}

type CreateCalendarManagementRequest struct {
	CountryID    string    `json:"country_id"`
	Date         time.Time `json:"date"`
	Descriptions string    `json:"descriptions"`
}

func (m CreateCalendarManagementRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.CountryID, validation.Required),
		validation.Field(&m.Date, validation.Required),
		validation.Field(&m.Descriptions, validation.Required),
	)
}

type UpdateCalendarManagementRequest struct {
	CountryID    string    `json:"country_id"`
	Date         time.Time `json:"date"`
	Descriptions string    `json:"descriptions"`
	Status       int8      `json:"status"`
}

func (m UpdateCalendarManagementRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.CountryID, validation.Required),
		validation.Field(&m.Date, validation.Required),
		validation.Field(&m.Descriptions, validation.Required),
	)
}
