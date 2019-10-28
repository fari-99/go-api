package routes

import (
	"go-api/configs"
	"os"

	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/recover"
)

type Routes struct {
	irisApp *iris.Application

	DB    *gorm.DB
	Redis *redis.Client
}

func NewRouteBase() *Routes {
	routes := &Routes{}

	// setup iris application
	routes.irisApp = configs.GetIrisApplication()

	// setup database
	routes.setupDatabase()

	// setup redis
	routes.setupRedis()

	return routes
}

func (routes *Routes) Setup(host string, port string) {
	app := routes.irisApp

	// recover from any http-relative panics
	app.Use(recover.New())

	// Set logging level
	app.Logger().SetLevel(os.Getenv("LOG_LEVEL"))

	routes.setupCustomerRoute()

	//start server
	_ = app.Run(iris.Addr(host+":"+port), iris.WithoutServerError(iris.ErrServerClosed))
}

func (routes *Routes) setupDatabase() *Routes {
	db := configs.DatabaseBase().GetDBConnection()

	// DB setup
	db.LogMode(false)

	// put db to routes
	routes.DB = db
	return routes
}

func (routes *Routes) setupRedis() *Routes {
	// Setup Redis
	redisConn := configs.GetRedis()

	// put redis to routes
	routes.Redis = redisConn
	return routes
}
