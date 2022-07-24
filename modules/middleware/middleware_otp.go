package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"go-api/constant"
	"go-api/helpers"
	"go-api/helpers/crypts"
	"go-api/modules/configs"
	"go-api/modules/models"

	"github.com/dgryski/dgoogauth"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func OTPMiddleware() gin.HandlerFunc {
	return otpServe
}

func otpServe(ctx *gin.Context) {
	uuid, _ := ctx.Get("uuid")
	currentUser, _ := helpers.GetCurrentUser(uuid.(string))

	twoAuthModel, err := getUserTwoAuthenticationModel(currentUser.ID)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	cryptBase := crypts.NewEncryptionBase()
	cryptBase.SetPassphrase(os.Getenv("2FA_KEY_ENCRYPT"))
	secret, err := cryptBase.Decrypt([]byte(twoAuthModel.Secret))
	if err != nil {
		helpers.NewResponse(ctx, http.StatusUnauthorized, err.Error())
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
}

func getUserTwoAuthenticationModel(userID uint64) (*models.TwoAuths, error) {
	db := configs.DatabaseBase(configs.MySQLType).GetMysqlConnection()

	var twoAuthModel models.TwoAuths
	err := db.Where(&models.TwoAuths{UserID: userID, Status: constant.StatusActive}).First(&twoAuthModel).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("please setup your two auth notification first, or your configuration is not found")
	} else if err != nil {
		return nil, err
	}

	return &twoAuthModel, nil
}
