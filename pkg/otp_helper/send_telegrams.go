package otp_helper

import (
	"gorm.io/gorm"

	"go-api/modules/models"
)

type telegramTemplate struct {
	Subject string
	Body    string
}

func getTelegramOtpTemplate(db *gorm.DB, action string) (*telegramTemplate, error) {
	var telegramTemplateModel models.NotificationTemplates
	err := db.Where(&models.NotificationTemplates{
		NotificationType: 4, // telegram
		Action:           action,
		Status:           99,
	}).First(&telegramTemplateModel).Error
	if err != nil {
		return nil, err
	}

	body := telegramTemplateModel.Body

	return &telegramTemplate{
		Subject: telegramTemplateModel.Subject,
		Body:    body,
	}, nil
}
