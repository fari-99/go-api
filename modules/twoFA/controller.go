package twoFA

import (
	"bytes"
	"fmt"
	"github.com/dgryski/dgoogauth"
	"github.com/gin-gonic/gin"
	"go-api/helpers"
	"io"
	"net/http"
	"rsc.io/qr"
)

type controller struct {
	service Service
}

func (c controller) CreateNewAuth(ctx *gin.Context) {
	_, notFound, err := c.service.GetDetails(ctx)
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

	img := code.PNG()
	buf := bytes.NewReader(img)

	responseWriter := ctx.Writer
	responseWriter.Header().Set("Content-Type", "image/png")
	responseWriter.WriteHeader(http.StatusOK)
	_, _ = io.Copy(responseWriter, buf)
	return
}

func (c controller) ValidateAuth(ctx *gin.Context) {
	twoAuthModel, notFound, err := c.service.GetDetails(ctx)
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
	_, notFound, err := c.service.GetDetails(ctx)
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

	code, err := c.service.GenerateRecoveryCode(ctx)
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
