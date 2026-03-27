package otp_helper

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"
	"time"

	"go-api/helpers"
	"go-api/helpers/notifications"
	"go-api/modules/models"

	"gorm.io/gorm"
)

type sendEmailRequest struct {
	db      *gorm.DB
	emailTo string
}

func sendEmailOtp(db *gorm.DB) *sendEmailRequest {
	return &sendEmailRequest{
		db: db,
	}
}

func (s *sendEmailRequest) setSendTo(userID uint64) *sendEmailRequest {
	db := s.db

	var userModel models.Users
	db.Where("id = ?", userID).First(&userModel)

	s.emailTo = userModel.Email
	return s
}

type emailNotificationData struct {
	OTP string
}

func (s *sendEmailRequest) send(action, otp string) error {
	emailData := notifications.Email{
		From: os.Getenv("EMAIL_OTP_FROM"),
		To:   []string{s.emailTo},
	}

	notificationData := emailNotificationData{
		OTP: otp,
	}

	err := s.generateBody(action, &emailData, &notificationData)
	if err != nil {
		return err
	}

	err = notifications.SendEmail(emailData)

	return err
}

func (s *sendEmailRequest) generateBody(action string, emailData *notifications.Email, notificationData *emailNotificationData) error {
	emailOtpTemplate, err := getEmailOtpTemplate(s.db, action)
	if err != nil {
		return err
	}

	funcMap := template.FuncMap{
		// The name "inc" is what the function will be called in the template text.
		"inc": func(i int) int {
			return i + 1
		},
		"string_date_format": func(date interface{}, parseFormat string, format string) string {
			formattedDate := ""

			switch date := date.(type) {
			case string:
				parsedDate, _ := time.Parse(parseFormat, date)
				formattedDate, err = helpers.ToLocale(&parsedDate, format)
				if err != nil {
					log.Println("error time ToLocale", err)
				}
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

	// compile body template
	bodyTemplate := template.New("body-notification-logs").Funcs(funcMap)
	_, err = bodyTemplate.Parse(emailOtpTemplate.Body)
	if err != nil {
		return fmt.Errorf("error parse body template, err := %s", err.Error())
	}

	bodyBuffer := bytes.NewBufferString("")
	err = bodyTemplate.Execute(bodyBuffer, notificationData)
	if err != nil {
		return fmt.Errorf("error execute data to body template, err := %s", err.Error())
	}

	// compile subject template
	subjectTemplate := template.New("subject-notification-logs")
	_, err = subjectTemplate.Parse(emailOtpTemplate.Body)
	if err != nil {
		return fmt.Errorf("error parse subject template, err := %s", err.Error())
	}

	subjectBuffer := bytes.NewBufferString("")
	err = subjectTemplate.Execute(subjectBuffer, notificationData)
	if err != nil {
		return fmt.Errorf("error execute data to subject template, err := %s", err.Error())
	}

	emailData.Subject = subjectBuffer.String()
	emailData.Body = bodyBuffer.String()
	return nil
}

type emailTemplate struct {
	Subject string
	Body    string
}

func getEmailOtpTemplate(db *gorm.DB, action string) (*emailTemplate, error) {
	var emailHeadersModel models.NotificationTemplates
	err := db.Where(&models.NotificationTemplates{
		NotificationType: 1, // email
		Action:           "otp_headers",
		Status:           99,
	}).First(&emailHeadersModel).Error
	if err != nil {
		return nil, err
	}

	var emailFooterModel models.NotificationTemplates
	err = db.Where(&models.NotificationTemplates{
		NotificationType: 1, // email
		Action:           "otp_footers",
		Status:           99,
	}).First(&emailFooterModel).Error
	if err != nil {
		return nil, err
	}

	var emailBodyModel models.NotificationTemplates
	err = db.Where(&models.NotificationTemplates{
		NotificationType: 1, // email
		Action:           action,
		Status:           99,
	}).First(&emailBodyModel).Error
	if err != nil {
		return nil, err
	}

	body := emailHeadersModel.Body + emailBodyModel.Body + emailFooterModel.Body

	return &emailTemplate{
		Subject: emailBodyModel.Subject,
		Body:    body,
	}, nil
}
