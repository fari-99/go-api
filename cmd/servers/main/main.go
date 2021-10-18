package main

import (
	"flag"
	"fmt"
	"go-api/modules/auths"
	"go-api/modules/configs"
	"go-api/modules/middleware"
	"go-api/modules/state_machine"
	"go-api/modules/storages"
	"go-api/modules/telegrams"
	"go-api/modules/twoFA"
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
	di := configs.DIInit()
	authentication := middleware.AuthMiddleware(middleware.BaseMiddleware{})
	//otpMiddleware := middleware.OTPMiddleware(middleware.BaseMiddleware{})

	auths.NewRegistrator(app.Group(""),
		auths.NewService(auths.NewRepository(di)))

	state_machine.NewRegistrator(app.Group(""),
		state_machine.NewService(state_machine.NewRepository(di)),
		authentication)

	storages.NewRegistrator(app.Group(""),
		storages.NewService(storages.NewRepository(di)),
		authentication)

	telegrams.NewRegistrator(app.Group(""),
		telegrams.NewService(telegrams.NewRepository(di)),
		authentication)

	twoFA.NewRegistrator(app.Group(""),
		twoFA.NewService(twoFA.NewRepository(di)),
		authentication)

	users.NewRegistrator(app.Group(""),
		users.NewService(users.NewRepository(di)),
		authentication)

	applicationRun := fmt.Sprintf("%s:%s", host, port)
	log.Printf("Run application on %s", applicationRun)
	_ = app.Run(applicationRun)
}
