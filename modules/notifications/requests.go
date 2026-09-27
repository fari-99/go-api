package notifications

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"

	"go-api/constant"
)

type RequestListFilter struct {
	Page             int    `json:"page"`
	Limit            int    `json:"limit"`
	OrderBy          string `json:"order_by"`
	NotificationType int8   `json:"notification_type" form:"notification_type"`
	Action           string `json:"action" form:"action"`
}

type RequestNotificationTemplate struct {
	NotificationType int8   `json:"notification_type"`
	Action           string `json:"action"`
	Subject          string `json:"subject"`
	Body             string `json:"body"`
	Status           int8   `json:"status"`
}

func (request RequestNotificationTemplate) Validate() error {
	return validation.ValidateStruct(&request,
		validation.Field(&request.NotificationType, validation.Required,
			validation.In(
				int8(constant.NotificationTypeEmail),
				int8(constant.NotificationTypeTelegram),
				int8(constant.NotificationTypeWhatsapp),
				int8(constant.NotificationTypeSMS),
				int8(constant.NotificationTypePushNotification),
			)),
		validation.Field(&request.Action, validation.Required),
		validation.Field(&request.Body, validation.Required),
	)
}

// RequestSendNotification manually sends an existing notification template to a single user.
type RequestSendNotification struct {
	UserID     int64 `json:"user_id"`
	TemplateID int64 `json:"template_id"`
}

func (request RequestSendNotification) Validate() error {
	return validation.ValidateStruct(&request,
		validation.Field(&request.UserID, validation.Required),
		validation.Field(&request.TemplateID, validation.Required),
	)
}
