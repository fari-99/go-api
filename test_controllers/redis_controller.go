package test_controllers

import (
	"fmt"
	"go-api/configs"

	"github.com/go-redis/redis"

	"github.com/jinzhu/gorm"
	"github.com/kataras/iris/v12"
)

type TestRedisController struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func (controller *TestRedisController) TestRedisAction(ctx iris.Context) {
	client := controller.Redis

	err := client.Set("key", "value", 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := client.Get("key").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("key", val)

	val2, err := client.Get("key2").Result()
	if err == redis.Nil {
		fmt.Println("key2 does not exist")
	} else if err != nil {
		panic(err)
	} else {
		fmt.Println("key2", val2)
	}
	_, _ = configs.NewResponse(ctx, iris.StatusOK, "yee")
	return
}
