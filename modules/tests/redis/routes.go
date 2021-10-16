package redis

import (
	"github.com/gin-gonic/gin"
	"go-api/modules/configs"
	"log"
)

func NewRoute(app *gin.Engine) {
	log.Println("Setup Test Redis router")

	testRedis := app.Group("/test-redis")
	{
		testRedisController := &TestRedisController{
			DI: configs.DIInit(),
		}

		testRedis.POST("/", testRedisController.TestRedisAction)
	}
}
