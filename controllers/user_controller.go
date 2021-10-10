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
	"os"
	"time"
)

type UserController struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func (controller *UserController) CreateAction(ctx *gin.Context) {
	db := controller.DB
	var input models.Users
	err := ctx.BindJSON(&input)
	if err != nil {
		configs.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	var userModel models.Users
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

func (controller *UserController) AuthenticateAction(ctx *gin.Context) {
	var input models.Users
	err := ctx.BindJSON(&input)
	if err != nil {
		configs.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	db := controller.DB
	var userModel models.Users
	if db.Where(&models.Users{Email: input.Email}).Find(&userModel).RecordNotFound() {
		configs.NewResponse(ctx, http.StatusOK, "User not found")
		return
	}

	err = helpers.AuthenticatePassword(&userModel, input.Password)
	if err != nil {
		configs.NewResponse(ctx, http.StatusOK, err.Error())
		return
	}

	// generate JWT token
	tokenHelper := token_generator.NewJwt().SetCtx(ctx)
	tokenHelper, err = tokenHelper.SetClaim(userModel)
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

		UserID:        userModel.ID,
		UserDetails:   userModel,
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
		Domain:   os.Getenv("PROJECT_DOMAIN"),
		Expires:  time.Unix(token.AccessExpiredAt, 0),
		Secure:   false,
		HttpOnly: true,
	})

	configs.NewResponse(ctx, http.StatusOK, tokenCompiled)
	return
}

func (controller *UserController) UserDetailsAction(ctx *gin.Context) {
	userUuid, exists := ctx.Get("uuid")
	if !exists {
		configs.NewResponse(ctx, http.StatusOK, "User not login or authentication failed")
	}

	userModel, err := helpers.GetCurrentUser(userUuid.(string))
	if err != nil {
		configs.NewResponse(ctx, http.StatusOK, err.Error())
		return
	}

	configs.NewResponse(ctx, http.StatusOK, userModel)
	return
}
