package otp_helper

import (
	"go-api/modules/models"

	"gorm.io/gorm"
)

type smsTemplate struct {
	Subject string
	Body    string
}

func getSmsOtpTemplate(db *gorm.DB, action string) (*smsTemplate, error) {
	var smsTemplateModel models.NotificationTemplates
	err := db.Where(&models.NotificationTemplates{
		NotificationType: 2, // sms
		Action:           action,
		Status:           99,
	}).First(&smsTemplateModel).Error
	if err != nil {
		return nil, err
	}

	body := smsTemplateModel.Body

	return &smsTemplate{
		Subject: smsTemplateModel.Subject,
		Body:    body,
	}, nil
}
