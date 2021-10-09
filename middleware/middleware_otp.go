package middleware

import (
	"fmt"
	"github.com/dgryski/dgoogauth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go-api/configs"
	"go-api/constant"
	"go-api/helpers"
	"go-api/models"
	"net/http"
	"os"
)

func OTPMiddleware(config BaseMiddleware) gin.HandlerFunc {
	defaultConfig := DefaultConfig()

	return defaultConfig.otpServe
}

func (config *BaseMiddleware) otpServe(ctx *gin.Context) {
	uuid, _ := ctx.Get("uuid")
	currentUser, _ := helpers.GetCurrentUser(uuid.(string))

	twoAuthModel, err := getUserTwoAuthenticationModel(currentUser.ID)
	if err != nil {
		configs.NewResponse(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	cryptBase := helpers.NewEncryptionBase()
	cryptBase.SetPassphrase(os.Getenv("2FA_KEY_ENCRYPT"))
	secret, err := cryptBase.Decrypt([]byte(twoAuthModel.Secret))
	if err != nil {
		configs.NewResponse(ctx, http.StatusUnauthorized, err.Error())
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
		configs.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if !isAuth {
		configs.NewResponse(ctx, http.StatusUnauthorized, fmt.Sprintf("failed to authenticate, try again"))
		return
	}
}

func getUserTwoAuthenticationModel(userID int64) (*models.TwoAuths, error) {
	db := configs.DatabaseBase().GetDBConnection()

	var twoAuthModel models.TwoAuths
	err := db.Where(&models.TwoAuths{UserID: userID, Status: constant.StatusActive}).First(&twoAuthModel).Error
	if err != nil && gorm.IsRecordNotFoundError(err) {
		return nil, fmt.Errorf("please setup your two auth notification first, or your configuration is not found")
	} else if err != nil {
		return nil, err
	}

	return &twoAuthModel, nil
}