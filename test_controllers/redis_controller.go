package test_controllers

import (
	"fmt"
	"github.com/kataras/iris/v12/sessions"
	"go-api/configs"

	"github.com/jinzhu/gorm"
	"github.com/kataras/iris/v12"
)

type TestRedisController struct {
	DB    *gorm.DB
	Redis *sessions.Sessions
}

func (controller *TestRedisController) TestRedisAction(ctx iris.Context) {
	redisConfig := controller.Redis
	sessionTest := redisConfig.Start(ctx)

	sessionTest.Set("key", "value")

	val := sessionTest.Get("key")
	fmt.Println("key", val)

	val2 := sessionTest.Get("key2")
	if val2 == nil {
		fmt.Println("key2 does not exist")
	}

	_, _ = configs.NewResponse(ctx, iris.StatusOK, "yee")
	return
}
