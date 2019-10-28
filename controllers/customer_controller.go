package controllers

import (
	"encoding/json"
	"fmt"
	"go-api/configs"
	"go-api/helpers"
	"go-api/helpers/token_generator"
	"go-api/models"

	"github.com/go-redis/redis"

	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

type CustomerController struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func (controller *CustomerController) CreateAction(ctx iris.Context) {
	db := controller.DB
	var input models.Customers
	err := ctx.ReadJSON(&input)
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusBadRequest, err.Error())
		return
	}

	var userModel models.Customers
	if !db.Debug().Where("username = ? OR email = ?", input.Username, input.Email).Find(&userModel).RecordNotFound() {
		_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, "Username or Email already created")
		return
	}

	password, err := helpers.GeneratePassword(input.Password)
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, err.Error())
		return
	}

	input.Password = password

	err = db.Create(&input).Error
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, err.Error())
		return
	}

	_, _ = configs.NewResponse(ctx, iris.StatusOK, "User successfully created")
	return
}

func (controller *CustomerController) AuthenticateAction(ctx iris.Context) {
	var input models.Customers
	err := ctx.ReadJSON(&input)
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusBadRequest, err.Error())
		return
	}

	db := controller.DB
	var customerModel models.Customers
	if db.Where(&models.Customers{Email: input.Email}).Find(&customerModel).RecordNotFound() {
		_, _ = configs.NewResponse(ctx, iris.StatusOK, "User not found")
		return
	}

	err = helpers.AuthenticatePassword(&customerModel, input.Password)
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusOK, err.Error())
		return
	}

	// generate JWT token
	token, err := token_generator.NewJwt().SetClaim(customerModel).SignClaim()
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, err.Error())
		return
	}

	var customerResult models.CustomerResult
	dataMarshal, _ := json.Marshal(customerModel)
	_ = json.Unmarshal(dataMarshal, &customerResult)

	customerResult.BearerToken = token

	_, _ = configs.NewResponse(ctx, iris.StatusOK, customerResult)
	return
}

func (controller *CustomerController) TestRedisAction(ctx iris.Context) {
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
