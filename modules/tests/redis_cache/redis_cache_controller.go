package redis_cache

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-api/helpers"
	"go-api/modules/configs"
	"net/http"
	"time"

	"github.com/go-redis/cache/v9"
)

type RedisCacheController struct {
	*configs.DI
}

type CacheObject struct {
	Str string
	Num int
}

func (controller *RedisCacheController) TestRedisCacheAction(ctx *gin.Context) {
	codec := controller.RedisCache

	key := "mykey"
	obj := &CacheObject{
		Str: "mystring",
		Num: 42,
	}

	_ = codec.Set(&cache.Item{
		Key:   key,
		Value: obj,
		TTL:   time.Hour,
	})

	var wanted CacheObject
	if err := codec.Get(ctx, key, &wanted); err == nil {
		fmt.Println(wanted)
	}

	helpers.NewResponse(ctx, http.StatusOK, "yee")
	return
}
