package main

import (
	"flag"
	"fmt"
	"go-api/modules/configs"
	"go-api/modules/state_machine"
	"go-api/modules/storages"
	"go-api/modules/telegrams"
	"go-api/modules/tokens"
	"go-api/modules/users"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	//get parameter from cli
	var host, port string
	flag.StringVar(&host, "host", os.Getenv("APP_HOST"), "host of the service")
	flag.StringVar(&port, "port", os.Getenv("GO_API_PORT"), "port of the service")
	flag.Parse()

	//info version service
	fmt.Printf("Service: %s\nVersion: %s\nParams:\n-host: host of the service\n-port: port of the service\nFramework:\n", os.Getenv("APP_NAME"), os.Getenv("APP_VER"))

	if rVal := recover(); rVal != nil {
		fmt.Printf("Rval: %+v\n", rVal)
	}

	// Setup routes and run application
	app := configs.GetGinApplication()
	state_machine.NewRoute(app)
	storages.NewRoute(app)
	telegrams.NewRoute(app)
	tokens.NewRoute(app)
	users.NewRoute(app)

	applicationRun := fmt.Sprintf("%s:%s", host, port)
	log.Printf("Run application on %s", applicationRun)
	_ = app.Run(applicationRun)
}
