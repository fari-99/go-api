package tasks

import (
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

type BaseCommand struct {
	*cli.Command
}

func NewBaseCommand() *BaseCommand {
	cmd := cli.Command{
		Version: "2.0.0",
		Name:    "Command execution for Go API CLI",
		Usage:   "Run task by command CLI for Golang",
		Authors: []any{
			"Fadhlan Rizal",
		},
	}

	baseCommand := &BaseCommand{
		&cmd,
	}

	return baseCommand
}

func (base *BaseCommand) GetFlags(cliCommand *cli.Command, flagName string) string {
	if cliCommand.NArg() > 0 {
		var dataArgs []string
		for i := 0; i < cliCommand.NArg(); i++ {
			dataArgs = append(dataArgs, cliCommand.Args().Get(i))
		}
		log.Printf("Your args := %+v", dataArgs)
	}

	return cliCommand.String(flagName)
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
