package routes

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"go-api/configs"
	"log"
	"os"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/go-redis/cache"
	"github.com/jinzhu/gorm"
	"gopkg.in/gomail.v2"
)

type Routes struct {
	ginApp *gin.Engine

	DB            *gorm.DB
	Redis         *redis.Client
	RedisCache    *cache.Codec
	Queue         *configs.QueueSetup
	EmailDialler  *gomail.Dialer
	ElasticSearch *elasticsearch.Client
}

func NewRouteBase() *Routes {
	routes := &Routes{}

	// setup gin application
	routes.ginApp = configs.GetGinApplication()
	return routes
}

func (routes *Routes) Setup(host string, port string) {
	app := routes.ginApp

	routes.setupCustomerRoute()
	routes.setupTokenRoute()
	routes.setupStorageRoute()

	if os.Getenv("LOG_LEVEL") == "debug" {
		routes.setupTestRoute()
	}

	applicationRun := fmt.Sprintf("%s:%s", host, port)
	log.Printf("Run application on %s", applicationRun)
	_ = app.Run(applicationRun)
}
