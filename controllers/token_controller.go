package controllers

import (
	"encoding/base64"
	"fmt"
	"go-api/configs"
	"go-api/helpers/token_generator"

	"github.com/jinzhu/gorm"

	"github.com/kataras/iris/v12"
)

type TokenController struct {
	DB *gorm.DB
}

func (controller *TokenController) CreateTokenAction(ctx iris.Context) {
	var input token_generator.AppData
	err := ctx.ReadJSON(&input)
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusBadRequest, err.Error())
		return
	}

	secretKey, err := controller.getSecretKeyApp(input.AppName)
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusBadRequest, err.Error())
		return
	}

	token, err := token_generator.NewJwt().SetSecretKey(secretKey).SetClaimApp(input).SignClaim()
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, err.Error())
		return
	}

	_, _ = configs.NewResponse(ctx, iris.StatusOK, iris.Map{
		"token_id": token,
	})
	return
}

func (controller *TokenController) getSecretKeyApp(appName string) (string, error) {
	appSecretKey := map[string]string{
		"test-company-name": base64.StdEncoding.EncodeToString([]byte("n^&4bZ@Y=WfQA!t2vxKU")),
	}

	if _, ok := appSecretKey[appName]; !ok {
		err := fmt.Errorf("app name not found")
		return "", err
	}

	return appSecretKey[appName], nil
}

func (controller *TokenController) CheckTokenAction(ctx iris.Context) {
	type InputCheck struct {
		AppName string `json:"app_name"`
		Token   string `json:"token"`
	}

	var input InputCheck
	err := ctx.ReadJSON(&input)
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusBadRequest, err.Error())
		return
	}

	secretKey, err := controller.getSecretKeyApp(input.AppName)
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusBadRequest, err.Error())
		return
	}

	claims, err := token_generator.NewJwt().SetSecretKey(secretKey).ParseToken(input.Token)
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusOK, iris.Map{
			"is_valid":      false,
			"error_message": err.Error(),
		})
		return
	}

	_, _ = configs.NewResponse(ctx, iris.StatusOK, iris.Map{
		"is_valid": true,
		"claims":   claims,
	})
	return
}
