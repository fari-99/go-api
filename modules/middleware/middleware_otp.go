package middleware

import (
	"fmt"
	"net/http"

	"github.com/dgryski/dgoogauth"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"

	"go-api/helpers"
	"go-api/modules/configs"
	"go-api/modules/twoFA"
)

type OtpConfig struct {
	DI *configs.DI
}

func OTPMiddlewareLogin(di *configs.DI) gin.HandlerFunc {
	otpConfig := OtpConfig{DI: di}
	return otpConfig.otpServe
}

func (otpConfig OtpConfig) otpServe(ctx *gin.Context) {
	uuid, _ := ctx.Get("uuid")
	currentUser, _ := helpers.GetCurrentUser(uuid.(string))
	userID := currentUser.ID

	if !currentUser.TwoFaEnabled {
		helpers.NewResponse(ctx, http.StatusUnauthorized, gin.H{
			"error_message": "Please enabled Two Factor Authorization to access this page",
		})
		ctx.Abort()
		return
	}

	twoFAService := twoFA.NewService(twoFA.NewRepository(otpConfig.DI))
	twoAuthModel, notFound, err := twoFAService.GetDetails(ctx, string(userID))
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "error get 2FA configuration for your user",
		})
		ctx.Abort()
		return
	} else if !notFound {
		helpers.NewResponse(ctx, http.StatusNotFound, gin.H{
			"error":         "user config found",
			"error_message": "your user already created the configuration, inactive the configuration first, and try create again",
		})
		ctx.Abort()
		return
	}

	secret, err := twoFAService.DecryptKey(*twoAuthModel)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusUnauthorized, gin.H{
			"error":         err.Error(),
			"error_message": "failed to decrypt key of your configuration",
		})
		ctx.Abort()
		return
	}

	recoveryCodeModels, err := twoFAService.GetAllRecoveryCode(ctx, string(twoAuthModel.UserID))
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to get your backup code",
		})
		ctx.Abort()
		return
	}

	var scratchCodes []int
	for _, recoveryCodeModel := range recoveryCodeModels {
		scratchCodes = append(scratchCodes, cast.ToInt(recoveryCodeModel.Code))
	}

	otpConfigs := &dgoogauth.OTPConfig{
		Secret:       string(secret),
		WindowSize:   3,
		HotpCounter:  0,
		ScratchCodes: scratchCodes,
	}

	otpValue := ctx.DefaultQuery("otp_value", "")
	isAuth, err := otpConfigs.Authenticate(otpValue)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		ctx.Abort()
		return
	}

	if !isAuth {
		helpers.NewResponse(ctx, http.StatusUnauthorized, fmt.Sprintf("failed to authenticate, try again"))
		ctx.Abort()
		return
	}

	ctx.Next()
}
