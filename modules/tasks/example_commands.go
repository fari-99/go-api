package tasks

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/fari-99/go-helper/rabbitmq"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/urfave/cli/v3"

	"go-api/helpers"
	"go-api/modules/configs"
	"go-api/modules/configs/kafka"
)

func (base *BaseCommand) getTestingCommands() []*cli.Command {
	command := []*cli.Command{
		{
			Name:        "queue-test-consumer",
			Aliases:     []string{"qtc"},
			Usage:       "queue-test-consumer --queue-name test",
			Description: "Testing queue consumer data",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "queue-name",
					Aliases:  []string{"qn"},
					Usage:    "queue name for this test",
					Required: true,
					Value:    "test-queue", // default value
				},
			},
			Action: func(ctx context.Context, command *cli.Command) error {
				log.Printf("====  Running task queue test consumer ====")
				queueName := base.GetFlags(command, "queue-name")
				log.Printf("Queue Name := %s", queueName)

				queueSetup := rabbitmq.NewBaseQueue("", queueName)
				queueSetup.SetupQueue(nil, nil)
				queueSetup.AddConsumer(false)
				queueSetup.Consume(HandleQueueEvents)

				queueSetup.WaitForSignalAndShutdown()
				log.Printf("====  Task success ====")
				return nil
			},
		},
		{
			Name:        "exchange-test-event",
			Aliases:     []string{"ete"},
			Usage:       "exchange-test-event",
			Description: "Exchange testing",
			Action: func(ctx context.Context, command *cli.Command) error {
				log.Printf("====  Running task exchange test consumer ====")

				queueSetup := rabbitmq.NewBaseQueue("", "test-queue")
				queueSetup.SetupExchange(nil)
				queueSetup.SetupQueue(nil, nil)
				queueSetup.SetupQueueBind(nil)
				queueSetup.AddConsumerExchange(false)
				queueSetup.Consume(HandleQueueEvents)

				queueSetup.WaitForSignalAndShutdown()
				return nil
			},
		},
		{
			Name:        "kafka-test-event",
			Aliases:     []string{"kte"},
			Usage:       "kafka-test-event",
			Description: "Kafka Consumer Testing",
			Action: func(ctx context.Context, command *cli.Command) error {
				log.Printf("====  Running Kafka Consumer Testing ====")

				kafkaConsumer, err := kafka.NewConsumer("consumerGroup", nil)
				if err != nil {
					panic(err)
				}

				consumerData := make(chan interface{}, 1)
				kafkaConsumer.Subscribe(kafka.NewConsumerHandler(consumerData))

				data := <-consumerData
				dataMarshal, err := json.Marshal(data)
				if err != nil {
					panic(err)
				}

				log.Printf(string(dataMarshal))

				return nil
			},
		},
		{
			Name:        "telegram-messages-commands",
			Aliases:     []string{"tm"},
			Usage:       "telegram-messages",
			Description: "Telegram Messages Handling",
			Action: func(ctx context.Context, command *cli.Command) error {
				bot := configs.GetTelegram()

				timeout, _ := strconv.ParseInt(os.Getenv("TELEGRAM_TIMEOUT"), 10, 64)

				updates := bot.GetUpdatesChan(tgbotapi.UpdateConfig{
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

func HandleQueueEvents(body rabbitmq.ConsumerHandlerData) {
	bodyMarshal, _ := json.Marshal(body)
	log.Printf(string(bodyMarshal))

	helpers.LoggingMessage("me sleep 25s now", nil)
	time.Sleep(25 * time.Second)
	helpers.LoggingMessage("done sleep", nil)

	testPanic := make([]int64, 0)
	log.Printf("%v", testPanic[100])
}
