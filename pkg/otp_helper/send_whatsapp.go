package otp_helper

import (
	"go-api/helpers/notifications"
	"go-api/modules/models"

	"gorm.io/gorm"
)

type sendWhatsappRequest struct {
	db    *gorm.DB
	phone string
}

func sendWhatsappOtp(db *gorm.DB) *sendWhatsappRequest {
	return &sendWhatsappRequest{
		db: db,
	}
}

func (s *sendWhatsappRequest) setSendTo(userID uint64) *sendWhatsappRequest {
	db := s.db

	var userModel models.Users
	db.Where("id = ?", userID).First(&userModel)

	s.phone = userModel.Phone
	return s
}

type whatsappNotificationData struct {
	OTP string
}

func (s *sendWhatsappRequest) send(action, otp string) error {
	whatsappData := notifications.WhatsappData{
		To: s.phone,
	}

	notificationData := whatsappNotificationData{
		OTP: otp,
	}

	err := s.generateBody(action, &whatsappData, &notificationData)
	if err != nil {
		return err
	}

	err = notifications.SendWhatsapp(whatsappData)

	return err
}

func (s *sendWhatsappRequest) generateBody(action string, whatsappData *notifications.WhatsappData, notificationData *whatsappNotificationData) error {
	whatsappOtpTemplate, err := getWhatsappOtpTemplate(s.db, action)
	if err != nil {
		return err
	}

	notifTemplateData := []notifications.NotificationTemplate{
		{
			NotificationType: 3, // whatsapp
			Subject:          whatsappOtpTemplate.Subject,
			Body:             whatsappOtpTemplate.Body,
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

	whatsappData.Message = notifResult[3].Body
	return nil
}

type whatsappTemplate struct {
	Subject string
	Body    string
}

func getWhatsappOtpTemplate(db *gorm.DB, action string) (*whatsappTemplate, error) {
	var smsTemplateModel models.NotificationTemplates
	err := db.Where(&models.NotificationTemplates{
		NotificationType: 3, // whatsapp
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
