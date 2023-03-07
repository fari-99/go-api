package handlers

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/twilio/twilio-go"
	"github.com/urfave/cli"
	"go.mau.fi/whatsmeow"
	"gorm.io/gorm"
)

type BaseEventHandler struct {
	DB               *gorm.DB
	Logger           log.Logger
	CliContext       *cli.Context
	Telegram         *tgbotapi.BotAPI
	TwilioRestClient *twilio.RestClient
	WhatsappClient   *whatsmeow.Client
}

type EventHandlerFlags struct {
	Limit      int64
	Worker     int64
	StatusType int64
}

func NewBaseEventHandler(transactionDB *gorm.DB, cliContext *cli.Context) *BaseEventHandler {
	baseEventHandler := BaseEventHandler{
		DB:         transactionDB,
		CliContext: cliContext,
	}

	return &baseEventHandler
}

func (base *BaseEventHandler) SetClientWhatsapp(client *whatsmeow.Client) *BaseEventHandler {
	base.WhatsappClient = client
	return base
}

func (base *BaseEventHandler) SetTelegram(telegramBotAPI *tgbotapi.BotAPI) *BaseEventHandler {
	base.Telegram = telegramBotAPI
	return base
}

func (base *BaseEventHandler) SetTwilio(twilioRestClient *twilio.RestClient) *BaseEventHandler {
	base.TwilioRestClient = twilioRestClient
	return base
}

func (base *BaseEventHandler) GetFlags() EventHandlerFlags {
	eventHandlerFlags := EventHandlerFlags{
		Limit:  10,
		Worker: 1000,
	}

	cliContext := base.CliContext

	if cliContext.NArg() > 0 {
		var dataArgs []string
		for i := 0; i < cliContext.NArg(); i++ {
			dataArgs = append(dataArgs, cliContext.Args().Get(i))
		}
	}

	return eventHandlerFlags
}
