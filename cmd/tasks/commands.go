package main

import (
	"context"
	"log"
	"os"
	"sort"

	"go-api/modules/tasks"

	"github.com/urfave/cli/v3"
)

func main() {
	app := tasks.NewBaseCommand()

	// populate command queue
	app.CommandQueueTask()
	sort.Sort(cli.FlagsByName(app.Flags))

	err := app.Run(context.Background(), os.Args)
	if err != nil {
		log.Printf(err.Error())
	}
}
