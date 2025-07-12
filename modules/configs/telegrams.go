package configs

import (
	"log"
	"os"
	"strconv"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramConfig struct {
	BotApi *tgbotapi.BotAPI
}

var TelegramInstance *TelegramConfig
var TelegramOnce sync.Once

func GetTelegram() *tgbotapi.BotAPI {
	TelegramOnce.Do(func() {
		log.Println("Initialize Telegram connection...")

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

		log.Println("Success Initialize Telegram connection...")
	})

	return TelegramInstance.BotApi
}
