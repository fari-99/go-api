package notifications

import (
	"bytes"
	"fmt"
	"go-api/constant/constant_models"
	"html/template"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type NotificationHelper struct {
	eventName string

	notificationTemplates []NotificationTemplate
}

type NotificationTemplate struct {
	NotificationType int8   `json:"notification_type"`
	Subject          string `json:"subject"`
	Body             string `json:"body"`
}

func (m NotificationTemplate) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.NotificationType, validation.Required),
		validation.Field(&m.Subject, validation.Required),
		validation.Field(&m.Body, validation.Required),
	)
}

type Notifications map[int8]CompiledNotification // key : notification type

type CompiledNotification struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func NewNotificationHelper(templatesData []NotificationTemplate) (*NotificationHelper, error) {
	base := NotificationHelper{}
	err := base.setNotificationTemplate(templatesData)
	if err != nil {
		return nil, err
	}

	return &base, nil
}

func (base *NotificationHelper) setNotificationTemplate(templatesData []NotificationTemplate) error {
	var notificationTemplates []NotificationTemplate
	for _, templateData := range templatesData {
		if err := templateData.Validate(); err != nil {
			return err
		}

		notificationTemplates = append(notificationTemplates, templateData)
	}

	base.notificationTemplates = notificationTemplates
	return nil
}

func (base *NotificationHelper) CompileNotificationTemplate(notificationData interface{}) (Notifications, error) {
	notificationTemplates := base.notificationTemplates
	notificationTypeLabels := constant_models.GetNotificationTypeLabel()

	notifications := make(Notifications)
	for _, notificationTemplate := range notificationTemplates {
		notificationType := notificationTemplate.NotificationType
		notificationTypeLabel := notificationTypeLabels[notificationType]

		// compile body template
		bodyTemplate := template.New(fmt.Sprintf("body-notification-%s-logs", notificationTypeLabel))
		_, err := bodyTemplate.Parse(notificationTemplate.Body)
		if err != nil {
			return nil, fmt.Errorf("error parse body template -%s-, err := %s", notificationTypeLabel, err.Error())
		}

		bodyBuffer := bytes.NewBufferString("")
		err = bodyTemplate.Execute(bodyBuffer, notificationData)
		if err != nil {
			return nil, fmt.Errorf("error execute data to body template -%s-, err := %s", notificationTypeLabel, err.Error())
		}

		// compile subject template
		subjectTemplate := template.New(fmt.Sprintf("subject-notification-%s-logs", notificationTypeLabel))
		_, err = subjectTemplate.Parse(notificationTemplate.Subject)
		if err != nil {
			return nil, fmt.Errorf("error parse subject template -%s-, err := %s", notificationTypeLabel, err.Error())
		}

		subjectBuffer := bytes.NewBufferString("")
		err = subjectTemplate.Execute(subjectBuffer, notificationData)
		if err != nil {
			return nil, fmt.Errorf("error execute data to subject template -%s-, err := %s", notificationTypeLabel, err.Error())
		}

		if _, ok := notifications[notificationType]; !ok {
			notifications[notificationType] = CompiledNotification{
				Subject: subjectBuffer.String(),
				Body:    bodyBuffer.String(),
			}
		}
	}

	return notifications, nil
}
