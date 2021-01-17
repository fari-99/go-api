package tasks

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"go-api/configs"
	"log"
	"os"
	"strconv"

	"github.com/urfave/cli"
)

func (base *BaseCommand) getQueueConsumerTask() []cli.Command {
	command := []cli.Command{
		{
			Name:        "queue-test-consumer",
			Aliases:     []string{"qtc"},
			Usage:       "queue-test-consumer --queue-name test",
			Description: "Testing queue consumer data",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "queue-name",
					Usage:    "queue name for this test",
					Required: true,
					Value:    "test-queue", // default value is 10
				},
			},
			Action: func(cliContext *cli.Context) (err error) {
				log.Printf("====  Running task queue test consumer ====")
				queueName := base.GetFlags(cliContext, "queue-name")
				log.Printf("Queue Name := %s", queueName)

				queueSetup := configs.NewBaseQueue().SetQueueName(queueName)

				configQueueDeclare := &configs.QueueDeclareConfig{
					Durable:    false,
					AutoDelete: false,
					Exclusive:  false,
					NoWait:     false,
					Args:       nil,
				}

				configConsumer := &configs.ConsumerConfig{
					Consumer:  "",
					AutoAck:   false,
					Exclusive: false,
					NoLocal:   false,
					NoWait:    false,
					Args:      nil,
				}

				queueSetup.AddConsumer(configQueueDeclare, configConsumer)
				queueSetup.Consume(HandleQueueEvents)

				log.Printf("====  Task success ====")
				return nil
			},
		},
		{
			Name:        "telegram-messages-commands",
			Aliases:     []string{"tm"},
			Usage:       "telegram-messages",
			Description: "Telegram Messages Handling",
			Action: func(cliContext *cli.Context) (err error) {
				bot := configs.GetTelegram()

				timeout, _ := strconv.ParseInt(os.Getenv("TELEGRAM_TIMEOUT"), 10, 64)

				updates, err := bot.GetUpdatesChan(tgbotapi.UpdateConfig{
					Offset:  0,
					Limit:   0,
					Timeout: int(timeout),
				})

				for update := range updates {
					if update.Message == nil {
						continue
					}

					dataMarshal, _ := json.MarshalIndent(update, "", " ")
					log.Printf(string(dataMarshal))

					log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

					if update.Message.IsCommand() {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
						switch update.Message.Command() {
						case "help":
							msg.Text = "type /sayhi or /status."
						case "sayhi":
							msg.Text = "Hi :)"
						case "status":
							msg.Text = "I'm ok."
						default:
							msg.Text = "I don't know that command"
						}

						_, _ = bot.Send(msg)
					}

				}

				return nil
			},
		},
	}

	return command
}

func HandleQueueEvents(body string) {
	log.Printf(fmt.Sprint(body))
}
