package notifications

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/cast"

	"go-api/modules/configs"
)

type TelegramData struct {
	Message string `json:"message"`
	To      string `json:"to"`
}

func SendTelegram(data TelegramData) error {
	client := configs.GetTelegram()
	msg := tgbotapi.NewMessage(cast.ToInt64(data.To), data.Message)
	_, err := client.Send(msg)
	return err
}
