package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"go-api/constant"
	"go-api/helpers/notifications"
	"go-api/modules/configs/rabbitmq"
	"go-api/modules/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/spf13/cast"

	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

type NotificationData struct {
	UsersSendTo          map[int64]models.Users `json:"users_send_to"`
	NotificationTemplate notifications.CompiledNotification
	ActionBy             *models.Users `json:"action_by"`
}

func (base *BaseEventHandler) NotificationEmailHandler(body rabbitmq.ConsumerHandlerData) {
	var input NotificationData
	dataMarshal, _ := json.Marshal(body.Data)
	_ = json.Unmarshal(dataMarshal, &input)

	var sentTo []string
	for _, userSendTo := range input.UsersSendTo {
		sentTo = append(sentTo, userSendTo.Email)
	}

	emailData := notifications.Email{
		Subject: input.NotificationTemplate.Subject,
		Body:    input.NotificationTemplate.Body,
		To:      sentTo,
	}

	if input.ActionBy == nil {
		emailData.From = os.Getenv("EMAIL_FROM_DEFAULT")
	} else {
		emailData.From = input.ActionBy.Email
	}

	err := notifications.SendEmail(emailData)
	if err != nil {
		log.Printf("error send email := %s\n", err.Error())
		return
	}

	return
}

func (base *BaseEventHandler) NotificationTelegramHandler(body rabbitmq.ConsumerHandlerData) {
	var input NotificationData
	dataMarshal, _ := json.Marshal(body.Data)
	_ = json.Unmarshal(dataMarshal, &input)

	for _, userSendTo := range input.UsersSendTo {
		if userSendTo.UserSocials == nil || len(userSendTo.UserSocials) < 1 {
			continue
		}

		var telegramToken string
		for _, userSocial := range userSendTo.UserSocials {
			if userSocial.NotificationType == constant.NotificationTypeTelegram {
				telegramToken = userSocial.Token
				break
			}
		}

		if telegramToken == "" {
			log.Printf("user [%d] not set telegram token", userSendTo.ID)
			continue
		}

		msg := tgbotapi.NewMessage(cast.ToInt64(telegramToken), input.NotificationTemplate.Body)
		_, err := base.Telegram.Send(msg)
		if err != nil {
			log.Printf("error send telegram [%d] := %s\n", userSendTo.ID, err.Error())
			continue
		}
	}

	return
}

func (base *BaseEventHandler) NotificationWhatsappHandler(body rabbitmq.ConsumerHandlerData) {
	var input NotificationData
	dataMarshal, _ := json.Marshal(body.Data)
	_ = json.Unmarshal(dataMarshal, &input)

	for _, userSendTo := range input.UsersSendTo {
		params := &openapi.CreateMessageParams{}
		params.SetTo(fmt.Sprintf("whatsapp:%s", userSendTo.MobilePhone))
		params.SetFrom(os.Getenv("TWILIO_NUMBER"))
		params.SetBody(input.NotificationTemplate.Body)

		_, err := base.TwilioRestClient.Api.CreateMessage(params)

		if err != nil {
			fmt.Printf("error send whatsapp [%s] := %s\n", userSendTo.ID, err.Error())
			continue
		}
	}

	return
}
