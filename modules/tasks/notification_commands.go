package tasks

import (
	"go-api/modules/configs"
	"go-api/modules/configs/rabbitmq"
	"go-api/modules/tasks/handlers"

	"github.com/urfave/cli"
)

func (base *BaseCommand) getNotificationCommands() []cli.Command {
	command := []cli.Command{
		{
			Name:        "handler-notifications-event",
			Aliases:     []string{"hne"},
			Usage:       "handler-notifications-event",
			Description: "Generate Notification template to send to respected notification type",
			Action: func(ctx *cli.Context) (err error) {
				db := configs.DatabaseBase().GetDBConnection()
				baseEvent := handlers.NewBaseEventHandler(db, ctx)

				queueSetup := rabbitmq.NewBaseQueue("", "notification-queue")
				queueSetup.SetupExchange(nil)
				queueSetup.SetupQueue(nil, nil)
				queueSetup.SetupQueueBind(nil)
				queueSetup.AddConsumerExchange(false)
				queueSetup.Consume(baseEvent.NotificationsHandler)
				return
			},
		},
		{
			Name:        "handler-notification-email-event",
			Aliases:     []string{"hnee"},
			Usage:       "handler-notification-email-event",
			Description: "Send generated template using email",
			Action: func(ctx *cli.Context) (err error) {
				db := configs.DatabaseBase().GetDBConnection()
				botAPI := configs.GetTelegram()
				twilioInit := configs.GetTwilioRestClient()

				baseEvent := handlers.NewBaseEventHandler(db, ctx)
				baseEvent.SetTelegram(botAPI)
				baseEvent.SetTwilio(twilioInit)

				queueSetup := rabbitmq.NewBaseQueue("", "email-queue")
				queueSetup.SetupQueue(nil, nil)
				queueSetup.AddConsumer(false)
				queueSetup.Consume(baseEvent.NotificationsHandler)
				return
			},
		},
		{
			Name:        "handler-notification-telegram-event",
			Aliases:     []string{"hnte"},
			Usage:       "handler-notification-telegram-event",
			Description: "Send generated template using telegram",
			Action: func(ctx *cli.Context) (err error) {
				db := configs.DatabaseBase().GetDBConnection()
				botAPI := configs.GetTelegram()
				twilioInit := configs.GetTwilioRestClient()

				baseEvent := handlers.NewBaseEventHandler(db, ctx)
				baseEvent.SetTelegram(botAPI)
				baseEvent.SetTwilio(twilioInit)

				queueSetup := rabbitmq.NewBaseQueue("", "telegram-queue")
				queueSetup.SetupQueue(nil, nil)
				queueSetup.AddConsumer(false)
				queueSetup.Consume(baseEvent.NotificationsHandler)
				return
			},
		},
		{
			Name:        "handler-notification-whatsapp-event",
			Aliases:     []string{"hnwe"},
			Usage:       "handler-notification-whatsapp-event",
			Description: "Send generated template using whatsapp",
			Action: func(ctx *cli.Context) (err error) {
				db := configs.DatabaseBase().GetDBConnection()
				botAPI := configs.GetTelegram()
				twilioInit := configs.GetTwilioRestClient()

				baseEvent := handlers.NewBaseEventHandler(db, ctx)
				baseEvent.SetTelegram(botAPI)
				baseEvent.SetTwilio(twilioInit)

				queueSetup := rabbitmq.NewBaseQueue("", "whatsapp-queue")
				queueSetup.SetupQueue(nil, nil)
				queueSetup.AddConsumer(false)
				queueSetup.Consume(baseEvent.NotificationsHandler)
				return
			},
		},
	}

	return command
}
