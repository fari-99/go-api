package tokens

import (
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-api/helpers"
	"go-api/helpers/token_generator"
	"go-api/modules/configs"
	"net/http"
)

type TokenController struct {
	*configs.DI
}

func (controller *TokenController) CreateTokenAction(ctx *gin.Context) {
	var input token_generator.AppData
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	secretKey, err := controller.getSecretKeyApp(input.AppName)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	token, err := token_generator.NewJwt().SetSecretKey(secretKey).SetClaimApp(input).SignClaims()
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, gin.H{
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

func (controller *TokenController) CheckTokenAction(ctx *gin.Context) {
	type InputCheck struct {
		AppName string `json:"app_name"`
		Token   string `json:"token"`
	}

	var input InputCheck
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	secretKey, err := controller.getSecretKeyApp(input.AppName)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	claims, err := token_generator.NewJwt().SetSecretKey(secretKey).ParseToken(input.Token)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusOK, gin.H{
			"is_valid":      false,
			"error_message": err.Error(),
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, gin.H{
		"is_valid": true,
		"claims":   claims,
	})
	return
}
