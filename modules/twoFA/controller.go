package twoFA

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go-api/pkg/otp_helper"

	"github.com/dgryski/dgoogauth"
	"github.com/fari-99/go-helper/token_generator"
	"github.com/gin-gonic/gin"
	"github.com/go-redsync/redsync/v4"

	"go-api/helpers"
	"go-api/modules/configs"
	"go-api/pkg/redis_helpers"
)

type controller struct {
	service Service
}

func (c controller) disableAuthenticator(ctx *gin.Context, userID uint64) gin.H {
	_, notFound, err := c.service.GetDetails(ctx, userID)
	if err != nil {
		return gin.H{
			"error":         err.Error(),
			"error_message": "error get 2FA configuration for your user",
		}
	} else if notFound {
		return gin.H{
			"error":         "user config not found",
			"error_message": "user doesn't have 2FA configuration, please create one",
		}
	}

	var input Request2FADisabled
	err = ctx.BindJSON(&input)
	if err != nil {
		return gin.H{
			"error":         err.Error(),
			"error_message": "failed to bind json input",
		}
	}

	userModel, notFound, err := c.service.GetUserDetails(ctx, userID)
	if err != nil {
		return gin.H{
			"error":         err.Error(),
			"error_message": "error getting user details",
		}
	} else if notFound {
		return gin.H{
			"error":         "user not found",
			"error_message": "user not found",
		}
	}

	err = helpers.PasswordAuth(userModel.Password, input.Password)
	if err != nil {
		return gin.H{
			"error":         err.Error(),
			"error_message": "wrong password",
		}
	}

	return nil
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

	secret, authLink, err := c.service.CreateTotp(ctx)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "error create 2FA [TOTP] configuration for your user",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, gin.H{
		"secret": secret,
		"link":   authLink,
	})
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

	errData := c.disableAuthenticator(ctx, userID)
	if errData != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, errData)
		return
	}

	err := c.service.UserEnabledTotp(ctx, userID, false)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"error":         err.Error(),
			"error_message": "failed to update 2FA configuration",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, fmt.Sprintf("success disable 2FA [TOTP]"))
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

func (c controller) DisableRecoveryCode(ctx *gin.Context) {
	uuid, _ := ctx.Get("uuid")
	currentUser, _ := helpers.GetCurrentUser(ctx, uuid.(string))
	userID := currentUser.ID.Uint64()

	errData := c.disableAuthenticator(ctx, userID)
	if errData != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, errData)
		return
	}

	err := c.service.DeleteAllRecoveryCodes(ctx, userID)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"error":         err.Error(),
			"error_message": "failed to update 2FA configuration",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, gin.H{
		"message": "successfully disabled 2FA configuration [Recovery Code]",
	})
	return
}

func (c controller) EnabledOtp(ctx *gin.Context) {
	uuid, _ := ctx.Get("uuid")
	senderType, _ := ctx.Params.Get("sender_type")

	currentUser, _ := helpers.GetCurrentUser(ctx, uuid.(string))
	userID := currentUser.ID.Uint64()

	// check if OTP with this sender type already enabled
	_, notFound, err := c.service.GetOtpDetails(ctx, userID, senderType)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "error checking existing OTP configuration",
		})
		return
	}
	if !notFound {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error_message": fmt.Sprintf("OTP via [%s] is already enabled, disable it first before re-enabling", senderType),
		})
		return
	}

	// create the OTP record
	err = c.service.CreateOtpRecord(ctx, userID, senderType)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"error":         err.Error(),
			"error_message": "failed to create OTP configuration",
		})
		return
	}

	// send initial OTP so the user can validate the setup
	otpSender := otp_helper.NewOtpSender(ctx.Request.Context()).SetUserID(userID)
	err = otpSender.SendOtp(senderType, "enable_otp")
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"error":         err.Error(),
			"error_message": fmt.Sprintf("OTP record created but failed to send OTP via [%s]", senderType),
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, gin.H{
		"message": fmt.Sprintf("OTP setup initiated via [%s], please validate the OTP sent to you", senderType),
	})
}

func (c controller) CreateOtp(ctx *gin.Context) {
	uuid, _ := ctx.Get("uuid")
	senderType, _ := ctx.Params.Get("sender_type")
	action, _ := ctx.Params.Get("action")

	currentUser, _ := helpers.GetCurrentUser(ctx, uuid.(string))
	userID := currentUser.ID.Uint64()

	// ensure OTP config exists for this sender type
	_, notFound, err := c.service.GetOtpDetails(ctx, userID, senderType)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "error checking OTP configuration",
		})
		return
	}
	if notFound {
		helpers.NewResponse(ctx, http.StatusNotFound, gin.H{
			"error_message": fmt.Sprintf("OTP via [%s] is not enabled, please enable it first", senderType),
		})
		return
	}

	// redis lock: prevent concurrent OTP sends for the same user+sender+action
	keyRedLock := fmt.Sprintf("CREATE_OTP:%d:%s:%s", userID, senderType, action)
	redLock := configs.GetRedisLock()
	mutex := redLock.NewMutex(keyRedLock)
	if err = mutex.Lock(); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to acquire redis lock",
		})
		return
	}
	defer func(mutex *redsync.Mutex) {
		if ok, err := mutex.Unlock(); !ok || err != nil {
			helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
				"error_message": "failed to unlock redis lock for otp",
			})
			return
		}
	}(mutex)

	// counter: rate-limit OTP sends
	countRed := redis_helpers.CounterConfig{
		Ctx:    ctx.Request.Context(),
		UserID: userID,
		Type:   redis_helpers.CounterTypeOtp,
		Action: action,
	}
	if err = countRed.Count(); err != nil {
		helpers.NewResponse(ctx, http.StatusTooManyRequests, gin.H{
			"error":         err.Error(),
			"error_message": "too many OTP requests, please try again later",
		})
		return
	}

	otpSender := otp_helper.NewOtpSender(ctx.Request.Context()).SetUserID(userID)
	if err = otpSender.SendOtp(senderType, action); err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"error":         err.Error(),
			"error_message": fmt.Sprintf("failed to send OTP via [%s]", senderType),
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, gin.H{
		"message": fmt.Sprintf("OTP sent via [%s]", senderType),
	})
}

func (c controller) ValidateOtp(ctx *gin.Context) {
	uuid, _ := ctx.Get("uuid")
	senderType, _ := ctx.Params.Get("sender_type")
	action, _ := ctx.Params.Get("action")

	currentUser, _ := helpers.GetCurrentUser(ctx, uuid.(string))
	userID := currentUser.ID.Uint64()

	_, notFound, err := c.service.GetOtpDetails(ctx, userID, senderType)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "error checking OTP configuration",
		})
		return
	}
	if notFound {
		helpers.NewResponse(ctx, http.StatusNotFound, gin.H{
			"error_message": fmt.Sprintf("OTP via [%s] is not enabled", senderType),
		})
		return
	}

	// redis lock: prevent concurrent OTP validations for the same user+sender+action
	keyRedLock := fmt.Sprintf("VALIDATE_OTP:%d:%s:%s", userID, senderType, action)
	redLock := configs.GetRedisLock()
	mutex := redLock.NewMutex(keyRedLock)
	if err = mutex.Lock(); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to acquire redis lock",
		})
		return
	}
	defer func(mutex *redsync.Mutex) {
		if ok, err := mutex.Unlock(); !ok || err != nil {
			helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
				"error_message": "failed to unlock redis lock for otp",
			})
			return
		}
	}(mutex)

	// counter: rate-limit OTP validation attempts
	countRed := redis_helpers.CounterConfig{
		Ctx:    ctx.Request.Context(),
		UserID: userID,
		Type:   redis_helpers.CounterTypeOtp,
		Action: action,
	}
	if err = countRed.Count(); err != nil {
		helpers.NewResponse(ctx, http.StatusTooManyRequests, gin.H{
			"error":         err.Error(),
			"error_message": "too many OTP attempts, please try again later",
		})
		return
	}
	defer func() {
		_ = countRed.Reset()
	}()

	var input RequestValidateOtp
	if err = ctx.ShouldBindJSON(&input); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "invalid request body",
		})
		return
	}

	otpSender := otp_helper.NewOtpSender(ctx.Request.Context()).SetUserID(userID)
	if err = otpSender.VerifyOtp(senderType, action, input.OtpValue); err != nil {
		helpers.NewResponse(ctx, http.StatusUnauthorized, gin.H{
			"error":         err.Error(),
			"error_message": "OTP validation failed",
		})
		return
	}

	// on the enable_otp action, mark the user's 2FA as enabled
	if action == "enable_otp" {
		if err = c.service.UserEnabledTotp(ctx, userID, true); err != nil {
			helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
				"error":         err.Error(),
				"error_message": "OTP validated but failed to enable 2FA on user",
			})
			return
		}
	}

	helpers.NewResponse(ctx, http.StatusOK, gin.H{
		"message": "OTP validated successfully",
	})
}

func (c controller) DisableOtp(ctx *gin.Context) {
	uuid, _ := ctx.Get("uuid")
	senderType, _ := ctx.Params.Get("sender_type")

	currentUser, _ := helpers.GetCurrentUser(ctx, uuid.(string))
	userID := currentUser.ID.Uint64()

	// verify password before disabling
	errData := c.disableAuthenticator(ctx, userID)
	if errData != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, errData)
		return
	}

	_, notFound, err := c.service.GetOtpDetails(ctx, userID, senderType)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "error checking OTP configuration",
		})
		return
	}
	if notFound {
		helpers.NewResponse(ctx, http.StatusNotFound, gin.H{
			"error_message": fmt.Sprintf("OTP via [%s] is not enabled", senderType),
		})
		return
	}

	if err = c.service.DeleteOtpRecord(ctx, userID, senderType); err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"error":         err.Error(),
			"error_message": "failed to disable OTP configuration",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, gin.H{
		"message": fmt.Sprintf("OTP via [%s] has been disabled", senderType),
	})
}
