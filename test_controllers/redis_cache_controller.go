package test_controllers

import (
	"fmt"
	"go-api/configs"
	"time"

	"github.com/go-redis/cache"
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

type RedisCacheController struct {
	DB         *gorm.DB
	RedisCache *cache.Codec
}

type CacheObject struct {
	Str string
	Num int
}

func (controller *RedisCacheController) TestRedisCacheAction(ctx iris.Context) {
	codec := controller.RedisCache

	key := "mykey"
	obj := &CacheObject{
		Str: "mystring",
		Num: 42,
	}

	_ = codec.Set(&cache.Item{
		Key:        key,
		Object:     obj,
		Expiration: time.Hour,
	})

	var wanted CacheObject
	if err := codec.Get(key, &wanted); err == nil {
		fmt.Println(wanted)
	}

	_, _ = configs.NewResponse(ctx, iris.StatusOK, "yee")
	return
}
