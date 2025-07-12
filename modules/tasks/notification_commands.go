package tasks

import (
	"context"
	"log"

	"go-api/constant"
	"go-api/modules/configs"
	"go-api/modules/configs/rabbitmq"
	"go-api/modules/tasks/handlers"

	"github.com/urfave/cli/v3"
)

func (base *BaseCommand) getNotificationCommands() []*cli.Command {
	command := []*cli.Command{
		{
			Name:        "handler-notifications-event",
			Aliases:     []string{"hne"},
			Usage:       "handler-notifications-event",
			Description: "Generate Notification template to send to respected notification type",
			Action: func(ctx context.Context, command *cli.Command) error {
				log.Printf("Handling Generate Notification Template")
				db := configs.DatabaseBase(configs.MySQLType).GetMysqlConnection(true)
				baseEvent := handlers.NewBaseEventHandler(db, ctx, command)

				queueSetup := rabbitmq.NewBaseQueue("", constant.QueueNotificationTemplate)
				queueSetup.SetupExchange(nil)
				queueSetup.SetupQueue(nil, nil)
				queueSetup.SetupQueueBind(nil)
				queueSetup.AddConsumerExchange(false)
				queueSetup.Consume(baseEvent.NotificationsHandler)

				queueSetup.WaitForSignalAndShutdown()
				return nil
			},
		},
		{
			Name:        "handler-notification-email-event",
			Aliases:     []string{"hnee"},
			Usage:       "handler-notification-email-event",
			Description: "Send generated template using email",
			Action: func(ctx context.Context, command *cli.Command) error {
				log.Printf("Handling Email Notification")
				db := configs.DatabaseBase(configs.MySQLType).GetMysqlConnection(true)
				baseEvent := handlers.NewBaseEventHandler(db, ctx, command)

				queueSetup := rabbitmq.NewBaseQueue("", constant.QueueNotificationEmail)
				queueSetup.SetupQueue(nil, nil)
				queueSetup.AddConsumer(false)
				queueSetup.Consume(baseEvent.NotificationEmailHandler)

				queueSetup.WaitForSignalAndShutdown()
				return nil
			},
		},
		{
			Name:        "handler-notification-telegram-event",
			Aliases:     []string{"hnte"},
			Usage:       "handler-notification-telegram-event",
			Description: "Send generated template using telegram",
			Action: func(ctx context.Context, command *cli.Command) error {
				log.Printf("Handling Telegram Notification")
				db := configs.DatabaseBase(configs.MySQLType).GetMysqlConnection(true)
				botAPI := configs.GetTelegram()

				baseEvent := handlers.NewBaseEventHandler(db, ctx, command)
				baseEvent.SetTelegram(botAPI)

				queueSetup := rabbitmq.NewBaseQueue("", constant.QueueNotificationTelegram)
				queueSetup.SetupQueue(nil, nil)
				queueSetup.AddConsumer(false)
				queueSetup.Consume(baseEvent.NotificationTelegramHandler)

				queueSetup.WaitForSignalAndShutdown()
				return nil
			},
		},
		{
			Name:        "handler-notification-twilio-event",
			Aliases:     []string{"hnte"},
			Usage:       "handler-notification-twilio-event",
			Description: "Send generated template using whatsapp",
			Action: func(ctx context.Context, command *cli.Command) error {
				log.Printf("Handling Twilio Notification")
				db := configs.DatabaseBase(configs.MySQLType).GetMysqlConnection(true)
				twilioInit := configs.GetTwilioRestClient()

				baseEvent := handlers.NewBaseEventHandler(db, ctx, command)
				baseEvent.SetTwilio(twilioInit)

				queueSetup := rabbitmq.NewBaseQueue("", constant.QueueNotificationWhatsapp)
				queueSetup.SetupQueue(nil, nil)
				queueSetup.AddConsumer(false)
				queueSetup.Consume(baseEvent.NotificationTwilioHandler)

				queueSetup.WaitForSignalAndShutdown()
				return nil
			},
		},
		{
			Name:        "handler-notification-whatsapp-event",
			Aliases:     []string{"hnwe"},
			Usage:       "handler-notification-whatsapp-event",
			Description: "Send generated template using whatsapp",
			Action: func(ctx context.Context, command *cli.Command) error {
				log.Printf("Handling Whatsapp Notification")
				db := configs.DatabaseBase(configs.MySQLType).GetMysqlConnection(true)
				redis := configs.GetRedisSessionConfig()
				client := configs.WhatsappClient(ctx, redis)

				baseEvent := handlers.NewBaseEventHandler(db, ctx, command)
				baseEvent.SetClientWhatsapp(client)
				queueSetup := rabbitmq.NewBaseQueue("", constant.QueueNotificationWhatsapp)
				queueSetup.SetupQueue(nil, nil)
				queueSetup.AddConsumer(false)
				queueSetup.Consume(baseEvent.NotificationWhatsappHandler)

				queueSetup.WaitForSignalAndShutdown()
				return nil
			},
		},
	}

	return command
}
