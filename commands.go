package main

import (
	"go-api/tasks"
	"log"
	"os"
	"sort"

	"github.com/urfave/cli"
)

func main() {
	app := tasks.NewBaseCommand()

	// populate command queue
	app.CommandQueueTask()

	// populate command exchange

	// populate command task [task-name]

	sort.Sort(cli.CommandsByName(app.Commands))
	sort.Sort(cli.FlagsByName(app.Flags))

	err := app.Run(os.Args)
	if err != nil {
		log.Printf(err.Error())
	}
}
