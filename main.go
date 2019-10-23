package main

import (
	"flag"
	"fmt"
	"goService/configs"
	"goService/utils"
	"log"
	"os"

	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
)

func main() {
	//get parameter from cli
	var host, port string
	flag.StringVar(&host, "host", os.Getenv("HOST"), "host of the service")
	flag.StringVar(&port, "port", os.Getenv("PORT"), "port of the service")
	flag.Parse()

	db := utils.DatabaseBase().GetDBConnection()
	defer func(db *gorm.DB) {
		err := db.Close()
		if err != nil {
			log.Printf("Error setup DB := " + err.Error())
		}
	}(db)

	// DB setup
	db.LogMode(false)

	// Setup routes
	routes := &configs.Routes{DB: db}

	// Setup Redis
	redisConn := utils.GetRedis()
	defer func(redisConn *redis.Client) {
		err := redisConn.Close()
		if err != nil {
			log.Printf("Error setup Redis := " + err.Error())
		}
	}(redisConn)

	//info version service
	fmt.Printf("Service: %s\nVersion: %s\nParams:\n-host: host of the service\n-port: port of the service\nFramework:\n", os.Getenv("APP_NAME"), os.Getenv("APP_VER"))

	if rval := recover(); rval != nil {
		fmt.Printf("Rval: %+v\n", rval)
	}
	// Setup routes and run application
	routes.Setup(host, port)
}
