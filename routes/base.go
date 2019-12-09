package routes

import (
	"go-api/configs"
	"go-api/helpers"
	"os"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/go-redis/cache"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/recover"
	"gopkg.in/gomail.v2"
)

type Routes struct {
	irisApp *iris.Application

	DB            *gorm.DB
	Redis         *redis.Client
	RedisCache    *cache.Codec
	Queue         *configs.QueueSetup
	EmailDialler  *gomail.Dialer
	ElasticSearch *elasticsearch.Client
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

	// setup ElasticSearch Connection
	routes.setupElasticSearch()

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

func (routes *Routes) setupDatabase() *Routes {
	helpers.LoggingMessage("Setup configuration database", nil)
	db := configs.DatabaseBase().GetDBConnection()

	// put db to routes
	routes.DB = db
	return routes
}

func (routes *Routes) setupRedis() *Routes {
	helpers.LoggingMessage("Setup configuration redis", nil)

	// Setup Redis
	redisConn := configs.GetRedisSession()

	// put redis to routes
	routes.Redis = redisConn
	return routes
}

func (routes *Routes) setupRedisCache() *Routes {
	helpers.LoggingMessage("Setup configuration redis cache", nil)

	// Setup Redis Cache
	redisCache := configs.GetRedisCache()

	// put redis cache to routes
	routes.RedisCache = redisCache
	return routes
}

func (routes *Routes) setupRabbitMqQueue() *Routes {
	helpers.LoggingMessage("Setup configuration RabbitMQ", nil)

	// setup connection RabbitMq queue
	queueBase := configs.NewBaseQueue()
	utils := queueBase.GetQueueUtil()

	// put rabbitMq queue connection to routes
	routes.Queue = utils
	return routes
}

func (routes *Routes) setupEmail() *Routes {
	helpers.LoggingMessage("Setup configuration Email", nil)

	dialer := configs.GetEmail()
	routes.EmailDialler = dialer
	return routes
}

func (routes *Routes) setupElasticSearch() *Routes {
	helpers.LoggingMessage("Setup configuration ElasticSearch", nil)

	elasticSearch := configs.GetElasticSearch()
	routes.ElasticSearch = elasticSearch
	return routes
}
