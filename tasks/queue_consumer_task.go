package tasks

import (
	"fmt"
	"github.com/urfave/cli"
	"go-api/configs"
	"go-api/tasks/task_collections"
	"log"
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
			Action:      task_collections.TelegramConsumerTask,
		},
	}

	return command
}

func HandleQueueEvents(body string) {
	log.Printf(fmt.Sprint(body))
}
