package routes

import (
	"go-api/test_controllers"
	"log"

	"github.com/kataras/iris"
)

func (routes *Routes) setupTestRoute() *iris.Application {
	app := routes.irisApp

	// Redis Test Endpoint collection
	app.PartyFunc("/test-redis", func(customers iris.Party) {
		log.Println("Setup Test Redis router")

		testRedisController := &test_controllers.TestRedisController{
			DB:    routes.DB,
			Redis: routes.Redis,
		}

		customers.Post("/", testRedisController.TestRedisAction)
	})

	// Redis Cache Test Endpoint collection
	app.PartyFunc("/test-redis-cache", func(customers iris.Party) {
		log.Println("Setup Test Redis Cache router")

		testRedisController := &test_controllers.RedisCacheController{
			DB:         routes.DB,
			RedisCache: routes.RedisCache,
		}

		customers.Post("/", testRedisController.TestRedisCacheAction)
	})

	// Redis Cache Test Endpoint collection
	app.PartyFunc("/test-rabbit", func(customers iris.Party) {
		log.Println("Setup Test RabbitMq Connection router")

		testRabbitMqQueueController := &test_controllers.RabbitMqController{
			RabbitMqConnection: routes.RabbitQueue.Connection,
			RabbitMqChannel:    routes.RabbitQueue.Channel,
		}

		customers.Post("/queue", testRabbitMqQueueController.TestPublishQueueAction)
		customers.Post("/batch-queue", testRabbitMqQueueController.TestBatchPublishQueueAction)
	})

	return app
}
