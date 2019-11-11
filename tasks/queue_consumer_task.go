package tasks

import (
	"fmt"
	"go-api/helpers/queue"
	"log"

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
					Value:    "test", // default value is 10
				},
			},
			Action: func(cliContext *cli.Context) (err error) {
				log.Printf("====  Running task queue test consumer ====")
				consumerBase := queue.BaseConsumer()

				configQueue := consumerBase.GetDefaultConfigQueueDeclare()
				configConsumer := consumerBase.GetDefaultConfigConsumer()

				queueName := base.GetFlags(cliContext, "queue-name")
				log.Printf("Queue Name := %s", queueName)

				configQueue.Name = queueName
				configConsumer.QueueName = configQueue.Name

				consumerBase.SetConfigQueue(configQueue).SetConfigConsumer(configConsumer)
				err = consumerBase.Consume(HandleQueueEvents)
				if err != nil {
					log.Printf("Error consume task, err := %s", err.Error())
					return err
				}

				log.Printf("====  Task success ====")
				return nil
			},
		},
	}

	return command
}

func HandleQueueEvents(body map[string]interface{}) {
	log.Printf(fmt.Sprint(body))
}
