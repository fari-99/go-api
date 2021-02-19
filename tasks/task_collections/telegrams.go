package task_collections

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jinzhu/gorm"
	"github.com/urfave/cli"
	"go-api/configs"
	"go-api/models"
	"os"
	"strconv"
	"strings"
)

func TelegramConsumerTask(cliContext *cli.Context) (err error) {
	bot := configs.GetTelegram()

	timeout, _ := strconv.ParseInt(os.Getenv("TELEGRAM_TIMEOUT"), 10, 64)

	updates, err := bot.GetUpdatesChan(tgbotapi.UpdateConfig{
		Offset:  0,
		Limit:   0,
		Timeout: int(timeout),
	})

	if err != nil {
		panic(err.Error())
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "help":
				msg.Text = "type /sayhi or /status."
			case "sayhi":
				msg.Text = "Hi :)"
			case "status":
				msg.Text = "I'm ok."
			case "start":
				textSplit := strings.Split(update.Message.Text, " ")
				if len(textSplit) <= 1 {
					msg.Text = "wrong input, please try again"
					_, _ = bot.Send(msg)
					continue
				}

				db := configs.DatabaseBase().GetDBConnection()
				var socialModel models.SocialMedia
				err = db.Where(&models.SocialMedia{Uuid: textSplit[1]}).First(&socialModel).Error
				if err != nil && gorm.IsRecordNotFoundError(err) {
					msg.Text = "wrong code, please try again"
					_, _ = bot.Send(msg)
					continue
				} else if err != nil {
					msg.Text = "error get code associated with your account, please try again"
					_, _ = bot.Send(msg)
					continue
				} else if socialModel.TokenID != "" {
					msg.Text = "this user already registered to our website, if it's not you please contact our admin"
					_, _ = bot.Send(msg)
					continue
				}

				socialModel.TokenID = fmt.Sprintf("%d", update.Message.Chat.ID)
				err = db.Save(&socialModel).Error
				if err != nil {
					msg.Text = "failed bind your code, please try again using the same code"
					_, _ = bot.Send(msg)
					continue
				}

				msg.Text = "Success adding your user to our website, thank you for your support, please use /help for more command"
			default:
				msg.Text = "I don't know that command"
			}

			_, _ = bot.Send(msg)
		}

	}

	return nil
}
