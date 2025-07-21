package redis

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"go-api/helpers"
	"go-api/modules/configs"
)

type TestRedisController struct {
	*configs.DI
}

func (controller *TestRedisController) TestRedisAction(ctx *gin.Context) {
	redisConfig := controller.Redis
	expired := time.Unix(time.Now().Add(time.Minute*15).Unix(), 0)

	redisConfig.Set(ctx, "key", "value", expired.Sub(time.Now()))

	val, _ := redisConfig.Get(ctx, "key").Result()
	fmt.Println("key", val)

	val2, err := redisConfig.Get(ctx, "key2").Result()
	if err != nil {
		fmt.Println("key2 does not exist")
	} else {
		println(val2)
	}

	helpers.NewResponse(ctx, http.StatusOK, "yee")
	return
}
