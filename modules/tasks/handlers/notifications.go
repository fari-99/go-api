package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"go-api/constant"
	"go-api/constant/constant_models"
	"go-api/helpers/notifications"
	"go-api/modules/configs/rabbitmq"
	"go-api/modules/models"

	"gorm.io/gorm"
)

func (base *BaseEventHandler) NotificationsHandler(body rabbitmq.ConsumerHandlerData) {
	fmt.Printf("Handle event [%s]\n", body.EventType)
	var err error
	switch body.EventType {
	case constant.EventSendNotifications:
		err = compileNotificationTemplate(base.DB, body)
	default:
		fmt.Printf("there is no event [%s], ignoring event...\n", body.EventType)
	}

	if err != nil {
		fmt.Printf("error data task [%s] when consumed, err := %s", body.EventType, err.Error())
		return
	}

	fmt.Printf("Sucess handle event [%s]\n", body.EventType)
	return
}

func compileNotificationTemplate(db *gorm.DB, body rabbitmq.ConsumerHandlerData) error {
	type Input struct {
		Action   string `json:"action"`
		ActionBy string `json:"action_by"`
	}

	var input Input
	dataMarshal, _ := json.Marshal(body.Data)
	_ = json.Unmarshal(dataMarshal, &input)

	notificationTemplate, exists, err := getNotificationTemplate(db, input.Action)
	if err != nil {
		return err
	} else if !exists {
		return nil
	}

	notificationHelper, err := notifications.NewNotificationHelper(notificationTemplate)
	if err != nil {
		return err
	}

	// generate data needed for notifications
	data := map[string]interface{}{
		"example": "example",
	}

	compiledNotificationTemplates, err := notificationHelper.CompileNotificationTemplate(data)
	if err != nil {
		return err
	}

	// get send to
	usersSendTo := map[int64]models.Users{
		1: {Base: models.Base{ID: "1"}, Email: os.Getenv("EMAIL_FROM_DEFAULT"), UserSocials: []models.UserSocials{{Token: "123456789"}}},
	}

	// get action by
	actionBy, exists, err := getUserDetails(db, input.ActionBy)
	if err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("user [%d] not found", input.ActionBy)
	}

	notificationTypeLabels := constant_models.GetNotificationTypeLabel()

	for notificationType, compiledNotificationTemplate := range compiledNotificationTemplates {
		notificationTypeLabel := notificationTypeLabels[notificationType]

		var queueName string
		switch notificationType {
		case constant.NotificationTypeEmail:
			queueName = constant.QueueNotificationEmail
		case constant.NotificationTypeTelegram:
			queueName = constant.QueueNotificationTelegram
		case constant.NotificationTypeWhatsapp:
			queueName = constant.QueueNotificationWhatsapp
		case constant.NotificationTypeSMS:
		case constant.NotificationTypePushNotification:
		default:
			log.Printf("notification type [%s] not yet setup", notificationTypeLabel)
			continue
		}

		if queueName == "" {
			log.Printf("queue name not yet set for notification type %s", notificationTypeLabel)
			continue
		}

		queueData := NotificationData{
			UsersSendTo:          usersSendTo,
			NotificationTemplate: compiledNotificationTemplate,
			ActionBy:             actionBy,
		}

		queueDataMarshal, _ := json.Marshal(queueData)

		queueSetup := rabbitmq.NewBaseQueue("", queueName)
		queueSetup.SetupQueue(nil, nil)
		queueSetup.AddPublisher(nil, nil)
		err = queueSetup.Publish(string(queueDataMarshal))
		if err != nil {
			log.Printf("error publish notification event, err := %s", err.Error())
			continue
		}
	}

	log.Printf("success send notification template")
	return nil
}

func getUserDetails(db *gorm.DB, userID string) (*models.Users, bool, error) {
	if userID == "" {
		return nil, true, nil
	}

	var actionBy models.Users
	err := db.Where(&models.Users{Base: models.Base{ID: models.IDType(userID)}}).First(&actionBy).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}

	return &actionBy, true, nil
}

func getNotificationTemplate(db *gorm.DB, action string) ([]notifications.NotificationTemplate, bool, error) {
	var notificationTemplates []models.NotificationTemplates
	err := db.Where(&models.NotificationTemplates{Action: action, Status: constant.StatusActive}).Find(&notificationTemplates).Error
	if err != nil {
		return nil, false, err
	}

	if len(notificationTemplates) == 0 {
		return nil, false, nil
	}

	var notificationTemplatesData []notifications.NotificationTemplate
	for _, notificationTemplate := range notificationTemplates {
		notificationTemplateData := notifications.NotificationTemplate{
			NotificationType: notificationTemplate.NotificationType,
			Subject:          notificationTemplate.Subject,
			Body:             notificationTemplate.Body,
		}

		notificationTemplatesData = append(notificationTemplatesData, notificationTemplateData)
	}

	return notificationTemplatesData, true, nil
}
