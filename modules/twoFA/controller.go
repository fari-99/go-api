package twoFA

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/dgryski/dgoogauth"
	"github.com/fari-99/go-helper/token_generator"
	"github.com/gin-gonic/gin"
	"github.com/go-redsync/redsync/v4"
	"rsc.io/qr"

	"go-api/helpers"
	"go-api/modules/configs"
	"go-api/pkg/redis_helpers"
)

type controller struct {
	service Service
}

func (c controller) CreateTotp(ctx *gin.Context) {
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

	_, authLink, err := c.service.CreateTotp(ctx)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "error create 2FA [TOTP] configuration for your user",
		})
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

func (c controller) ValidateTotp(ctx *gin.Context) {
	uuid, _ := ctx.Get("uuid")
	action, _ := ctx.Params.Get("action")

	currentUser, _ := helpers.GetCurrentUser(ctx, uuid.(string))
	userID := currentUser.ID.Uint64()

	twoFAModelsCtx, _ := ctx.Get("two_fa_models")
	var dataTwoFAModels token_generator.TwoFAModels
	_ = json.Unmarshal([]byte(twoFAModelsCtx.(string)), &dataTwoFAModels)

	if action == redis_helpers.TwoFAActionTypeCreate && dataTwoFAModels.TOTP {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error_message": "please disable your totp first",
		})
		return
	}

	keyRedLock := "VALIDATE_TOTP:" + currentUser.ID.String()
	redLock := configs.GetRedisLock()
	mutex := redLock.NewMutex(keyRedLock)
	if err := mutex.Lock(); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to acquire redis lock",
		})
		return
	}

	defer func(mutex *redsync.Mutex) {
		if ok, err := mutex.Unlock(); !ok || err != nil {
			helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
				"error_message": "failed to unlock redis lock for totp",
			})
			return
		}
	}(mutex)

	countRed := redis_helpers.CounterConfig{
		Ctx:    ctx.Request.Context(),
		UserID: userID,
		Type:   redis_helpers.CounterTypeTotp,
		Action: action,
	}

	err := countRed.Count()
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "counter is not valid, please try again later",
		})
		return
	}

	defer func() {
		_ = countRed.Reset()
	}()

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

	if action == redis_helpers.TwoFAActionTypeCreate {
		if !dataTwoFAModels.TOTP {
			err = c.service.UserEnabledTotp(ctx, userID, true)
			if err != nil {
				helpers.NewResponse(ctx, http.StatusInternalServerError, map[string]interface{}{
					"error":         err.Error(),
					"error_message": "failed to update 2FA configuration",
				})
				return
			}
		}
	}

	helpers.NewResponse(ctx, http.StatusOK, fmt.Sprintf("success to authenticate"))
	return
}

func (c controller) DisabledTotp(ctx *gin.Context) {
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

	err = c.service.UserEnabledTotp(ctx, userID, false)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, map[string]interface{}{
			"error":         err.Error(),
			"error_message": "failed to update 2FA configuration",
		})
	}

	helpers.NewResponse(ctx, http.StatusOK, fmt.Sprintf("success to authenticate"))
	return
}

func (c controller) CreateRecoveryCode(ctx *gin.Context) {
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

	code, err := c.service.CreateRecoveryCode(ctx, userID)
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

func (c controller) ValidateRecoveryCode(ctx *gin.Context) {
	uuid, _ := ctx.Get("uuid")
	currentUser, _ := helpers.GetCurrentUser(ctx, uuid.(string))
	userID := currentUser.ID.Uint64()

	keyRedLock := "VALIDATE_RECOVERY_CODE:" + currentUser.ID.String()
	redLock := configs.GetRedisLock()
	mutex := redLock.NewMutex(keyRedLock)
	if err := mutex.Lock(); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to acquire redis lock for recovery code",
		})
		return
	}

	defer func(mutex *redsync.Mutex) {
		if ok, err := mutex.Unlock(); !ok || err != nil {
			helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
				"error_message": "failed to unlock redis lock for recovery code",
			})
			return
		}
	}(mutex)

	action, _ := ctx.Params.Get("action")
	countRed := redis_helpers.CounterConfig{
		Ctx:    ctx.Request.Context(),
		UserID: userID,
		Type:   redis_helpers.CounterTypeRecoveryCode,
		Action: action,
	}

	err := countRed.Count()
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "counter is not valid, please try again later",
		})
		return
	}

	defer func() {
		_ = countRed.Reset()
	}()

	_, notFound, err := c.service.GetDetails(ctx, userID)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "error get 2FA [recovery-code] configuration for your user",
		})
		return
	} else if notFound {
		helpers.NewResponse(ctx, http.StatusNotFound, gin.H{
			"error":         "user config not found",
			"error_message": "user doesn't have 2FA [recovery-code] configuration, please create one",
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
