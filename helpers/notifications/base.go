package notifications

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"
	"time"

	"go-api/constant/constant_models"
	"go-api/helpers"

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
	funcMap := base.getFuncMap()

	notifications := make(Notifications)
	for _, notificationTemplate := range notificationTemplates {
		notificationType := notificationTemplate.NotificationType
		notificationTypeLabel := notificationTypeLabels[notificationType]

		// compile body template
		bodyTemplate := template.New(fmt.Sprintf("body-notification-%s-logs", notificationTypeLabel)).Funcs(funcMap)
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
		subjectTemplate := template.New(fmt.Sprintf("subject-notification-%s-logs", notificationTypeLabel)).Funcs(funcMap)
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

func (base *NotificationHelper) getFuncMap() template.FuncMap {
	return template.FuncMap{
		// The name "inc" is what the function will be called in the template text.
		"inc": func(i int) int {
			return i + 1
		},
		"string_date_format": func(date interface{}, parseFormat string, format string) string {
			formattedDate := ""

			switch date := date.(type) {
			case string:
				parsedDate, _ := time.Parse(parseFormat, date)
				formattedDateData, err := helpers.ToLocale(&parsedDate, format)
				if err != nil {
					log.Println("error time ToLocale", err)
				}

				formattedDate = formattedDateData
			}

			return formattedDate
		},
		"date_format": func(date interface{}, format string) string {
			formattedDate := ""
			switch date := date.(type) {
			case time.Time:
				formattedDate, _ = helpers.ToLocale(&date, format)
			}

			return formattedDate
		},
		"join": func(sep string, s []string) string {
			return strings.Join(s, sep)
		},
		"default_url": func(urlType string) string {
			switch urlType {
			case "asset_url":
				return os.Getenv("DMP_S3_ASSETS_URL")
			case "logo_url":
				return os.Getenv("EMAIL_NEW_LOGO_URL")
			case "banner_url":
				return os.Getenv("EMAIL_BANNER_URL")
			case "youtube_url":
				return os.Getenv("EMAIL_YOUTUBE_URL")
			case "youtube_image":
				return os.Getenv("EMAIL_YOUTUBE_IMAGE")
			case "facebook_url":
				return os.Getenv("EMAIL_FACEBOOK_URL")
			case "facebook_image":
				return os.Getenv("EMAIL_FACEBOOK_IMAGE")
			case "instagram_url":
				return os.Getenv("EMAIL_INSTAGRAM_URL")
			case "instagram_image":
				return os.Getenv("EMAIL_INSTAGRAM_IMAGE")
			case "help_email":
				return os.Getenv("EMAIL_HELP_EMAIL")
			case "APP_BUYER_URL":
				return os.Getenv("APP_BUYER_URL")
			default:
				return ""
			}
		},
	}
}
