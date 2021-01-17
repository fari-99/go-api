package routes

import (
	"github.com/kataras/iris/v12/sessions"
	"go-api/configs"
	"os"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/go-redis/cache"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/recover"
	"gopkg.in/gomail.v2"
)

type Routes struct {
	irisApp *iris.Application

	DB            *gorm.DB
	Redis         *sessions.Sessions
	RedisCache    *cache.Codec
	Queue         *configs.QueueSetup
	EmailDialler  *gomail.Dialer
	ElasticSearch *elasticsearch.Client
}

func NewRouteBase() *Routes {
	routes := &Routes{}

	// setup iris application
	routes.irisApp = configs.GetIrisApplication()
	return routes
}

func (routes *Routes) Setup(host string, port string) {
	app := routes.irisApp

	// recover from any http-relative panics
	app.Use(recover.New())

	// Set logging level
	app.Logger().SetLevel(os.Getenv("LOG_LEVEL"))

	routes.setupCustomerRoute()
	routes.setupTokenRoute()
	routes.setupStorageRoute()

	if os.Getenv("LOG_LEVEL") == "debug" {
		routes.setupTestRoute()
	}

	//start server
	_ = app.Run(iris.Addr(host+":"+port), iris.WithoutServerError(iris.ErrServerClosed))
}
