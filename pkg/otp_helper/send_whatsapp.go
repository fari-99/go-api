package otp_helper

import (
	"go-api/modules/models"

	"gorm.io/gorm"
)

type whatsappTemplate struct {
	Subject string
	Body    string
}

func getWhatsappOtpTemplate(db *gorm.DB, action string) (*whatsappTemplate, error) {
	var smsTemplateModel models.NotificationTemplates
	err := db.Where(&models.NotificationTemplates{
		NotificationType: 3, // sms
		Action:           action,
		Status:           99,
	}).First(&smsTemplateModel).Error
	if err != nil {
		return nil, err
	}

	body := smsTemplateModel.Body

	return &whatsappTemplate{
		Subject: smsTemplateModel.Subject,
		Body:    body,
	}, nil
}
