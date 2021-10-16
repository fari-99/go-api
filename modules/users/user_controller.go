package users

import (
	"github.com/gin-gonic/gin"
	"go-api/helpers"
	"go-api/helpers/token_generator"
	"go-api/modules/configs"
	"go-api/modules/models"
	"net/http"
	"os"
	"time"
)

type UserController struct {
	*configs.DI
}

func (controller *UserController) CreateAction(ctx *gin.Context) {
	db := controller.DB
	var input models.Customers
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	var userModel models.Customers
	if !db.Debug().Where("username = ? OR email = ?", input.Username, input.Email).Find(&userModel).RecordNotFound() {
		helpers.NewResponse(ctx, http.StatusInternalServerError, "Username or EmailDialler already created")
		return
	}

	password, err := helpers.GeneratePassword(input.Password)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	input.Password = password

	err = db.Create(&input).Error
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "User successfully created")
	return
}

func (controller *UserController) AuthenticateAction(ctx *gin.Context) {
	var input models.Customers
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	db := controller.DB
	var customerModel models.Customers
	if db.Where(&models.Customers{Email: input.Email}).Find(&customerModel).RecordNotFound() {
		helpers.NewResponse(ctx, http.StatusOK, "User not found")
		return
	}

	err = helpers.AuthenticatePassword(&customerModel, input.Password)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusOK, err.Error())
		return
	}

	// generate JWT token
	tokenHelper := token_generator.NewJwt().SetCtx(ctx)
	tokenHelper, err = tokenHelper.SetClaim(customerModel)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	token, err := tokenHelper.SignClaims()
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, err.Error())
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
		helpers.NewResponse(ctx, http.StatusInternalServerError, err.Error())
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

	helpers.NewResponse(ctx, http.StatusOK, tokenCompiled)
	return
}

func (controller *UserController) CustomerDetailsAction(ctx *gin.Context) {
	userUuid, exists := ctx.Get("uuid")
	if !exists {
		helpers.NewResponse(ctx, http.StatusOK, "Customer not login or authentication failed")
	}

	customerModel, err := helpers.GetCurrentUser(userUuid.(string))
	if err != nil {
		helpers.NewResponse(ctx, http.StatusOK, err.Error())
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, customerModel)
	return
}


