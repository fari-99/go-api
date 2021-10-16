package redis_cache

import (
	"github.com/gin-gonic/gin"
	"go-api/modules/configs"
	"log"
)

func NewRoute(app *gin.Engine) {
	log.Println("Setup Test Redis Cache router")

	redisCache := app.Group("/test-redis-cache")
	{
		testRedisController := &RedisCacheController{
			DI: configs.DIInit(),
		}

		redisCache.POST("/", testRedisController.TestRedisCacheAction)
	}
}
