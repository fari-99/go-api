package tasks

import (
	"context"
	"log"

	"go-api/constant"
	"go-api/modules/configs"
	"go-api/modules/file_managements"

	"github.com/fari-99/go-helper/rabbitmq"
	"github.com/urfave/cli/v3"
)

func (base *BaseCommand) getFileManagementCommands() []*cli.Command {
	command := []*cli.Command{
		{
			Name:        "handler-file-management-event",
			Aliases:     []string{"hfme"},
			Usage:       "handler-file-management-event",
			Description: "Process queued file managements (pdf summary, audio transcription, ...) with the third party service",
			Action: func(ctx context.Context, command *cli.Command) error {
				log.Printf("Handling File Management")
				db := configs.DatabaseBase(configs.MySQLType).GetMysqlConnection(true)
				worker := file_managements.NewWorker(
					file_managements.NewRepository(&configs.DI{DB: db}),
					file_managements.RegistryFromEnv())

				queueSetup := rabbitmq.NewBaseQueue("", constant.QueueFileManagement)
				queueSetup.SetupQueue(file_managements.QueueDeclareConfig(), nil)
				queueSetup.AddConsumer(false)
				queueSetup.Consume(worker.Handle)

				queueSetup.WaitForSignalAndShutdown()
				return nil
			},
		},
	}

	return command
}
