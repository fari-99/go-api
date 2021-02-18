package test_controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"go-api/configs"
	"net/http"
	"time"

	"github.com/jinzhu/gorm"
)

type TestRedisController struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func (controller *TestRedisController) TestRedisAction(ctx *gin.Context) {
	redisConfig := controller.Redis
	expired := time.Unix(time.Now().Add(time.Minute*15).Unix(), 0)

	redisConfig.Set("key", "value", expired.Sub(time.Now()))

	val, _ := redisConfig.Get("key").Result()
	fmt.Println("key", val)

	val2, err := redisConfig.Get("key2").Result()
	if err != nil {
		fmt.Println("key2 does not exist")
	} else {
		println(val2)
	}

	configs.NewResponse(ctx, http.StatusOK, "yee")
	return
}
