package main

import (
	"flag"
	"fmt"
	"go-api/routes"
	"os"
)

func main() {
	//get parameter from cli
	var host, port string
	flag.StringVar(&host, "host", os.Getenv("APP_HOST"), "host of the service")
	flag.StringVar(&port, "port", os.Getenv("GO_API_PORT"), "port of the service")
	flag.Parse()

	//info version service
	fmt.Printf("Service: %s\nVersion: %s\nParams:\n-host: host of the service\n-port: port of the service\nFramework:\n", os.Getenv("APP_NAME"), os.Getenv("APP_VER"))

	if rval := recover(); rval != nil {
		fmt.Printf("Rval: %+v\n", rval)
	}

	// Setup routes and run application
	routes.NewRouteBase().Setup(host, port)
}
