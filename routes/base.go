package routes

import (
	"go-api/configs"
	"gopkg.in/gomail.v2"
	"os"

	"github.com/streadway/amqp"

	"github.com/go-redis/cache"

	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/recover"
)

type Routes struct {
	irisApp *iris.Application

	DB          *gorm.DB
	Redis       *redis.Client
	RedisCache  *cache.Codec
	RabbitQueue struct {
		Connection *amqp.Connection
		Channel    *amqp.Channel
	}
	EmailDialler *gomail.Dialer
}

func NewRouteBase() *Routes {
	routes := &Routes{}

	// setup iris application
	routes.irisApp = configs.GetIrisApplication()

	// setup database
	routes.setupDatabase()

	// setup redis
	routes.setupRedis()

	// setup redis cache
	routes.setupRedisCache()

	// setup RabbitMq Connection
	routes.setupRabbitMqQueue()

	// setup EmailDialler Connection
	routes.setupEmail()

	return routes
}

func (routes *Routes) Setup(host string, port string) {
	app := routes.irisApp

	// recover from any http-relative panics
	app.Use(recover.New())

	// Set logging level
	app.Logger().SetLevel(os.Getenv("LOG_LEVEL"))

	routes.setupCustomerRoute()

	if os.Getenv("LOG_LEVEL") == "debug" {
		routes.setupTestRoute()
	}

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

func (routes *Routes) setupRedisCache() *Routes {
	// Setup Redis Cache
	redisCache := configs.GetRedisCache()

	// put redis cache to routes
	routes.RedisCache = redisCache
	return routes
}

func (routes *Routes) setupRabbitMqQueue() *Routes {
	// setup connection RabbitMq queue
	connection, channel := configs.GetRabbitQueue()

	// put rabbitMq queue connection to routes
	routes.RabbitQueue.Channel = channel
	routes.RabbitQueue.Connection = connection
	return routes
}

func (routes *Routes) setupEmail() *Routes {
	dialer := configs.GetEmail()
	routes.EmailDialler = dialer
	return routes
}
