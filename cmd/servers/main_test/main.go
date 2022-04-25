package main

import (
	"flag"
	"fmt"
	"go-api/modules/configs"
	"go-api/modules/tests/crypts"
	"go-api/modules/tests/emails"
	"go-api/modules/tests/finite_states"
	"go-api/modules/tests/ftps"
	"go-api/modules/tests/kafka"
	"go-api/modules/tests/rabbitmq"
	"go-api/modules/tests/redis"
	"go-api/modules/tests/redis_cache"
	"go-api/modules/tests/twofa"
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
	crypts.NewRoute(app)
	emails.NewRoute(app)
	finite_states.NewRoute(app)
	ftps.NewRoute(app)
	rabbitmq.NewRoute(app)
	redis.NewRoute(app)
	redis_cache.NewRoute(app)
	twofa.NewRoute(app)
	kafka.NewRoute(app)

	applicationRun := fmt.Sprintf("%s:%s", host, port)
	log.Printf("Run application on %s", applicationRun)
	_ = app.Run(applicationRun)
}
