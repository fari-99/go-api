package twoFA

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/dgryski/dgoogauth"
	"github.com/gin-gonic/gin"
	"rsc.io/qr"

	"go-api/helpers"
)

type controller struct {
	service Service
}

func (c controller) CreateNewAuth(ctx *gin.Context) {
	uuid, _ := ctx.Get("uuid")
	currentUser, _ := helpers.GetCurrentUser(ctx, uuid.(string))
	userID := currentUser.ID

	_, notFound, err := c.service.GetDetails(ctx, userID.Uint64())
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "error get 2FA configuration for your user",
		})
		return
	} else if !notFound {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         "user config found",
			"error_message": "your user already created the configuration, inactive the configuration first, and try create again",
		})
		return
	}

	_, authLink, err := c.service.CreateConfigs(ctx)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	code, err := qr.Encode(authLink, qr.H)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	// TODO: Set Redis to "TEMP_ENABLED_2FA:UserID"

	img := code.PNG()
	buf := bytes.NewReader(img)

	responseWriter := ctx.Writer
	responseWriter.Header().Set("Content-Type", "image/png")
	responseWriter.WriteHeader(http.StatusOK)
	_, _ = io.Copy(responseWriter, buf)
	return
}

func (c controller) ValidateAuth(ctx *gin.Context) {
	uuid, _ := ctx.Get("uuid")
	currentUser, _ := helpers.GetCurrentUser(ctx, uuid.(string))
	userID := currentUser.ID.Uint64()

	twoAuthModel, notFound, err := c.service.GetDetails(ctx, userID)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "error get 2FA configuration for your user",
		})
		return
	} else if notFound {
		helpers.NewResponse(ctx, http.StatusNotFound, gin.H{
			"error":         "user config not found",
			"error_message": "user doesn't have 2FA configuration, please create one",
		})
		return
	}

	secret, err := c.service.DecryptKey(*twoAuthModel)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusUnauthorized, gin.H{
			"error":         err.Error(),
			"error_message": "failed to decrypt key of your configuration",
		})
		return
	}

	otpConfig := &dgoogauth.OTPConfig{
		Secret:      string(secret),
		WindowSize:  3,
		HotpCounter: 0,
	}

	otpValue := ctx.DefaultQuery("otp_value", "")

	isAuth, err := otpConfig.Authenticate(otpValue)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, map[string]interface{}{
			"error":         err.Error(),
			"error_message": "failed to authenticate otp",
		})
		return
	}

	if !isAuth {
		helpers.NewResponse(ctx, http.StatusUnauthorized, map[string]interface{}{
			"error":         "not authorized",
			"error_message": "failed to authenticate, try again",
		})
		return
	}

	// TODO: CHECK Redis to "TEMP_ENABLED_2FA:UserID", if true, then update user model to 2FA enabled
	err = c.service.TwoFAUserUpdate(ctx, userID, true)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, map[string]interface{}{
			"error":         err.Error(),
			"error_message": "failed to update 2FA configuration",
		})
	}

	helpers.NewResponse(ctx, http.StatusOK, fmt.Sprintf("success to authenticate"))
	return
}

func (c controller) DisabledAuth(ctx *gin.Context) {
	uuid, _ := ctx.Get("uuid")
	currentUser, _ := helpers.GetCurrentUser(ctx, uuid.(string))
	userID := currentUser.ID.Uint64()

	_, notFound, err := c.service.GetDetails(ctx, userID)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "error get 2FA configuration for your user",
		})
		return
	} else if notFound {
		helpers.NewResponse(ctx, http.StatusNotFound, gin.H{
			"error":         "user config not found",
			"error_message": "user doesn't have 2FA configuration, please create one",
		})
		return
	}

	var input Request2FADisabled
	err = ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to bind json input",
		})
		return
	}

	// TODO: check password to disabled
	err = helpers.PasswordAuth("", input.Password)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "wrong password",
		})
		return
	}

	err = c.service.TwoFAUserUpdate(ctx, userID, false)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, map[string]interface{}{
			"error":         err.Error(),
			"error_message": "failed to update 2FA configuration",
		})
	}

	helpers.NewResponse(ctx, http.StatusOK, fmt.Sprintf("success to authenticate"))
	return
}

func (c controller) ValidateRecoveryCodeAuth(ctx *gin.Context) {
	uuid, _ := ctx.Get("uuid")
	currentUser, _ := helpers.GetCurrentUser(ctx, uuid.(string))
	userID := currentUser.ID.Uint64()

	_, notFound, err := c.service.GetDetails(ctx, userID)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "error get 2FA configuration for your user",
		})
		return
	} else if notFound {
		helpers.NewResponse(ctx, http.StatusNotFound, gin.H{
			"error":         "user config not found",
			"error_message": "user doesn't have 2FA configuration, please create one",
		})
		return
	}

	var input RequestValidateRecoveryCode
	err = ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	isAuth, err := c.service.ValidateRecoveryCode(ctx, input.RecoveryCode, userID)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if !isAuth {
		helpers.NewResponse(ctx, http.StatusUnauthorized, fmt.Sprintf("failed to authenticate, try again"))
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, fmt.Sprintf("success to authenticate"))
	return
}

func (c controller) GenerateRecoveryCode(ctx *gin.Context) {
	uuid, _ := ctx.Get("uuid")
	currentUser, _ := helpers.GetCurrentUser(ctx, uuid.(string))
	userID := currentUser.ID.Uint64()

	_, notFound, err := c.service.GetDetails(ctx, userID)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "error get 2FA configuration for your user",
		})
		return
	} else if notFound {
		helpers.NewResponse(ctx, http.StatusNotFound, gin.H{
			"error":         "user config not found",
			"error_message": "user doesn't have 2FA configuration, please create one",
		})
		return
	}

	code, err := c.service.GenerateRecoveryCode(ctx, userID)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "error generate recovery code",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, code)
	return
}
