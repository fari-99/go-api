package otp_helper

import (
	"fmt"

	"go-api/helpers/notifications"
	"go-api/modules/models"

	"gorm.io/gorm"
)

type sendSmsRequest struct {
	db    *gorm.DB
	phone string
}

func sendSmsOtp(db *gorm.DB) *sendSmsRequest {
	return &sendSmsRequest{
		db: db,
	}
}

func (s *sendSmsRequest) setSendTo(userID uint64) *sendSmsRequest {
	db := s.db

	var userModel models.Users
	db.Where("id = ?", userID).First(&userModel)

	s.phone = userModel.MobilePhone
	return s
}

type smsNotificationData struct {
	OTP string
}

func (s *sendSmsRequest) send(action, otp string) error {
	phoneNumber := s.phone
	if phoneNumber == "" {
		return fmt.Errorf("user phone number is empty [sms]")
	}

	smsData := notifications.SmsData{
		To: phoneNumber,
	}

	notificationData := smsNotificationData{
		OTP: otp,
	}

	err := s.generateBody(action, &smsData, &notificationData)
	if err != nil {
		return err
	}

	err = notifications.SendSms(smsData)

	return err
}

func (s *sendSmsRequest) generateBody(action string, smsData *notifications.SmsData, notificationData *smsNotificationData) error {
	smsOtpTemplate, err := getSmsOtpTemplate(s.db, action)
	if err != nil {
		return err
	}

	notifTemplateData := []notifications.NotificationTemplate{
		{
			NotificationType: 2, // sms
			Subject:          smsOtpTemplate.Subject,
			Body:             smsOtpTemplate.Body,
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

	smsData.Message = notifResult[2].Body
	return nil
}

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
