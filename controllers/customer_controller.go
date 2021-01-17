package controllers

import (
	"github.com/kataras/iris/v12/sessions"
	"go-api/configs"
	"go-api/helpers"
	"go-api/helpers/token_generator"
	"go-api/models"

	"github.com/jinzhu/gorm"
	"github.com/kataras/iris/v12"
)

type CustomerController struct {
	DB    *gorm.DB
	Redis *sessions.Sessions
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
		_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, "Username or EmailDialler already created")
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
	tokenHelper := token_generator.NewJwt().SetCtx(ctx)
	tokenHelper, err = tokenHelper.SetClaim(customerModel)
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, err.Error())
		return
	}

	token, err := tokenHelper.SignClaims()
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, err.Error())
		return
	}

	dataSession := helpers.SessionData{
		AccessUuid:  token.AccessUuid,
		RefreshUuid: token.RefreshUuid,

		UserID:        customerModel.ID,
		UserDetails:   customerModel,
		Authorization: true,
	}

	_ = helpers.SetRedisSession(dataSession, ctx)

	tokenCompiled := map[string]interface{}{
		"access_token":  token.AccessToken,
		"refresh_token": token.AccessToken,
	}

	_, _ = configs.NewResponse(ctx, iris.StatusOK, tokenCompiled)
	return
}

func (controller *CustomerController) CustomerDetailsAction(ctx iris.Context) {
	userUuid := ctx.Values().Get("uuid")
	customerModel, err := helpers.GetCurrentUser(userUuid.(string), ctx)
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusOK, err.Error())
		return
	}

	_, _ = configs.NewResponse(ctx, iris.StatusOK, customerModel)
	return
}

func (controller *CustomerController) RefreshTokenAction(ctx iris.Context) {

}
