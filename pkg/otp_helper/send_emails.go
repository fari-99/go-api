package otp_helper

import (
	"os"

	"go-api/constant"
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

	notifTemplateData := []notifications.NotificationTemplate{
		{
			NotificationType: constant.NotificationTypeEmail,
			Subject:          emailOtpTemplate.Subject,
			Body:             emailOtpTemplate.Body,
		},
	}

	notifHelper, err := notifications.NewNotificationHelper(notifTemplateData)
	if err != nil {
		return err
	}

	notifResult, err := notifHelper.CompileNotificationTemplate(notificationData)
	if err != nil {
		return err
	}

	emailData.Body = notifResult[constant.NotificationTypeEmail].Body
	emailData.Subject = notifResult[constant.NotificationTypeEmail].Subject
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
