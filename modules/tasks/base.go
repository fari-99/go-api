package tasks

import (
	"log"
	"os"

	"github.com/urfave/cli"
)

type BaseCommand struct {
	*cli.App
}

func NewBaseCommand() *BaseCommand {

	baseCommand := &BaseCommand{
		cli.NewApp(),
	}

	baseCommand.Name = "Command execution for Go API CLI"
	baseCommand.Usage = "Run task by command CLI for Golang"
	baseCommand.Author = "Fadhlan Rizal"
	baseCommand.Version = "1.0.0"

	return baseCommand
}

func (base *BaseCommand) GetFlags(cliContext *cli.Context, flagName string) string {
	if cliContext.NArg() > 0 {
		var dataArgs []string
		for i := 0; i < cliContext.NArg(); i++ {
			dataArgs = append(dataArgs, cliContext.Args().Get(i))
		}
		log.Printf("Your args := %+v", dataArgs)
	}

	return cliContext.String(flagName)
}

func (base *BaseCommand) CommandQueueTask() *BaseCommand {
	commands := base.Commands

	if os.Getenv("APP_STATE") == "test" {
		queueConsumerCommands := base.getTestingCommands()
		for _, queueConsumerCommand := range queueConsumerCommands {
			commands = append(commands, queueConsumerCommand)
		}
	}

	queueCommands := base.getNotificationCommands()
	for _, queueCommand := range queueCommands {
		commands = append(commands, queueCommand)
	}

	base.Commands = commands
	return base
}
