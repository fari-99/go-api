package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"go-api/configs"
	"go-api/helpers"
	"go-api/helpers/token_generator"
	"go-api/models"
	"net/http"
	"time"
)

type CustomerController struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func (controller *CustomerController) CreateAction(ctx *gin.Context) {
	db := controller.DB
	var input models.Customers
	err := ctx.BindJSON(&input)
	if err != nil {
		configs.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	var userModel models.Customers
	if !db.Debug().Where("username = ? OR email = ?", input.Username, input.Email).Find(&userModel).RecordNotFound() {
		configs.NewResponse(ctx, http.StatusInternalServerError, "Username or EmailDialler already created")
		return
	}

	password, err := helpers.GeneratePassword(input.Password)
	if err != nil {
		configs.NewResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	input.Password = password

	err = db.Create(&input).Error
	if err != nil {
		configs.NewResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	configs.NewResponse(ctx, http.StatusOK, "User successfully created")
	return
}

func (controller *CustomerController) AuthenticateAction(ctx *gin.Context) {
	var input models.Customers
	err := ctx.BindJSON(&input)
	if err != nil {
		configs.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	db := controller.DB
	var customerModel models.Customers
	if db.Where(&models.Customers{Email: input.Email}).Find(&customerModel).RecordNotFound() {
		configs.NewResponse(ctx, http.StatusOK, "User not found")
		return
	}

	err = helpers.AuthenticatePassword(&customerModel, input.Password)
	if err != nil {
		configs.NewResponse(ctx, http.StatusOK, err.Error())
		return
	}

	// generate JWT token
	tokenHelper := token_generator.NewJwt().SetCtx(ctx)
	tokenHelper, err = tokenHelper.SetClaim(customerModel)
	if err != nil {
		configs.NewResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	token, err := tokenHelper.SignClaims()
	if err != nil {
		configs.NewResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	dataSession := helpers.SessionData{
		Token: helpers.SessionToken{
			AccessExpiredAt:  token.AccessExpiredAt,
			AccessUuid:       token.AccessUuid,
			RefreshExpiredAt: token.RefreshExpiredAt,
			RefreshUuid:      token.RefreshUuid,
		},

		UserID:        customerModel.ID,
		UserDetails:   customerModel,
		Authorization: true,
	}

	err = helpers.SetRedisSession(dataSession)
	if err != nil {
		configs.NewResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	tokenCompiled := map[string]interface{}{
		"access_token":  token.AccessToken,
		"refresh_token": token.AccessToken,
	}

	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     "token",
		Value:    token.AccessToken,
		Path:     "/",
		Domain:   ".fadhlan.loc",
		Expires:  time.Unix(token.AccessExpiredAt, 0),
		Secure:   false,
		HttpOnly: true,
	})

	configs.NewResponse(ctx, http.StatusOK, tokenCompiled)
	return
}

func (controller *CustomerController) CustomerDetailsAction(ctx *gin.Context) {
	userUuid, exists := ctx.Get("uuid")
	if !exists {
		configs.NewResponse(ctx, http.StatusOK, "Customer not login or authentication failed")
	}

	customerModel, err := helpers.GetCurrentUser(userUuid.(string))
	if err != nil {
		configs.NewResponse(ctx, http.StatusOK, err.Error())
		return
	}

	configs.NewResponse(ctx, http.StatusOK, customerModel)
	return
}

func (controller *CustomerController) RefreshTokenAction(ctx *gin.Context) {

}
