package configs

import (
	"log"
	"os"
	"strconv"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type TelegramConfig struct {
	BotApi *tgbotapi.BotAPI
}

var TelegramInstance *TelegramConfig
var TelegramOnce sync.Once

func GetTelegram() *tgbotapi.BotAPI {
	TelegramOnce.Do(func() {
		bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_API_KEY"))
		if err != nil {
			log.Panic(err)
		}

		log.Printf("Authorized on account %s", bot.Self.UserName)

		isDebug, _ := strconv.ParseBool(os.Getenv("TELEGRAM_DEBUG"))
		bot.Debug = isDebug

		TelegramInstance = &TelegramConfig{
			BotApi: bot,
		}
	})

	return TelegramInstance.BotApi
}
