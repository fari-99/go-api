package routes

import (
	"go-api/test_controllers"
	"log"

	"github.com/kataras/iris"
)

func (routes *Routes) setupTestRoute() *iris.Application {
	app := routes.irisApp
	db := routes.DB
	redis := routes.Redis

	//authentication := configs.NewMiddleware(configs.MiddlewareConfiguration{})

	// Redis Test Endpoint collection
	app.PartyFunc("/test-redis", func(customers iris.Party) {
		log.Println("Setup Test Redis router")

		testRedisController := &test_controllers.TestRedisController{
			DB:    db,
			Redis: redis,
		}

		customers.Post("/", testRedisController.TestRedisAction)
	})

	return app
}
