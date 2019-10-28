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

	return app
}
