package otp_helper

import (
	"go-api/helpers/notifications"
	"go-api/modules/models"

	"gorm.io/gorm"
)

type sendTelegramRequest struct {
	db         *gorm.DB
	telegramID string
}

func sendTelegramOtp(db *gorm.DB) *sendTelegramRequest {
	return &sendTelegramRequest{
		db: db,
	}
}

func (s *sendTelegramRequest) setSendTo(userID uint64) *sendTelegramRequest {
	db := s.db

	var userModel models.Users
	db.Where("id = ?", userID).First(&userModel)

	s.telegramID = userModel.TelegramId
	return s
}

type telegramNotificationData struct {
	OTP string
}

func (s *sendTelegramRequest) send(action, otp string) error {
	telegramData := notifications.TelegramData{
		To: s.telegramID,
	}

	notificationData := telegramNotificationData{
		OTP: otp,
	}

	err := s.generateBody(action, &telegramData, &notificationData)
	if err != nil {
		return err
	}

	err = notifications.SendTelegram(telegramData)

	return err
}

func (s *sendTelegramRequest) generateBody(action string, telegramData *notifications.TelegramData, notificationData *telegramNotificationData) error {
	telegramOtpTemplate, err := getTelegramOtpTemplate(s.db, action)
	if err != nil {
		return err
	}

	notifTemplateData := []notifications.NotificationTemplate{
		{
			NotificationType: 4, // telegram
			Subject:          telegramOtpTemplate.Subject,
			Body:             telegramOtpTemplate.Body,
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

	telegramData.Message = notifResult[4].Body
	return nil
}

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
