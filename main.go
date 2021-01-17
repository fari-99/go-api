package main

import (
	"flag"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/go-redis/cache"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris/v12/sessions"
	"go-api/configs"
	"go-api/helpers"
	"go-api/routes"
	"gopkg.in/gomail.v2"
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

	if rVal := recover(); rVal != nil {
		fmt.Printf("Rval: %+v\n", rVal)
	}

	// Setup routes and run application
	routesSetup := routes.NewRouteBase()
	routesSetup.DB = setupDatabase()
	routesSetup.Redis = setupRedis()
	routesSetup.RedisCache = setupRedisCache()
	routesSetup.Queue = setupRabbitMqQueue()
	routesSetup.EmailDialler = setupEmail()
	routesSetup.ElasticSearch = setupElasticSearch()
	routesSetup.Setup(host, port)
}

func setupDatabase() *gorm.DB {
	helpers.LoggingMessage("Setup configuration database", nil)
	return configs.DatabaseBase().GetDBConnection()
}

func setupRedis() *sessions.Sessions {
	helpers.LoggingMessage("Setup configuration redis", nil)
	return configs.GetRedisSessionConfig()
}

func setupRedisCache() *cache.Codec {
	helpers.LoggingMessage("Setup configuration redis cache", nil)
	return configs.GetRedisCache()
}

func setupRabbitMqQueue() *configs.QueueSetup {
	helpers.LoggingMessage("Setup configuration RabbitMQ", nil)
	return configs.NewBaseQueue().GetQueueUtil()
}

func setupEmail() *gomail.Dialer {
	helpers.LoggingMessage("Setup configuration Email", nil)
	return configs.GetEmail()
}

func setupElasticSearch() *elasticsearch.Client {
	helpers.LoggingMessage("Setup configuration ElasticSearch", nil)
	return configs.GetElasticSearch()
}
