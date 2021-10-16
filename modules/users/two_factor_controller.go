package users

import (
	"bytes"
	"encoding/base32"
	"fmt"
	"github.com/dgryski/dgoogauth"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go-api/constant"
	"go-api/helpers"
	"go-api/helpers/crypts"
	"go-api/modules/configs"
	"go-api/modules/models"
	"io"
	"net/http"
	"os"
	"rsc.io/qr"
)

type TwoFactorAuthController struct {
	*configs.DI
}

func (controller *TwoFactorAuthController) getUserTwoAuthenticationModel(userID int64) (*models.TwoAuths, error) {
	var twoAuthModel models.TwoAuths
	err := controller.DB.Where(&models.TwoAuths{UserID: userID, Status: constant.StatusActive}).First(&twoAuthModel).Error
	if err != nil && gorm.IsRecordNotFoundError(err) {
		return nil, fmt.Errorf("please setup your two auth notification first, or your configuration is not found")
	} else if err != nil {
		return nil, err
	}

	return &twoAuthModel, nil
}

func (controller *TwoFactorAuthController) CreateNewAuth(ctx *gin.Context) {
	uuid, _ := ctx.Get("uuid")
	currentUser, _ := helpers.GetCurrentUser(uuid.(string))

	oldTwoAuthModel, _ := controller.getUserTwoAuthenticationModel(currentUser.ID)
	if oldTwoAuthModel != nil {
		oldTwoAuthModel.Status = constant.StatusNonActive
		err := controller.DB.Save(&oldTwoAuthModel).Error
		if err != nil {
			helpers.NewResponse(ctx, http.StatusBadRequest, fmt.Sprintf("failed to update old 2FA model status, err := %s", err.Error()))
			return
		}
	}

	account := currentUser.Email
	issuer := os.Getenv("APP_NAME")
	secret := crypts.GenerateRandString(10, "alphanum")
	encodedSecret := base32.StdEncoding.EncodeToString([]byte(secret))

	cryptBase := crypts.NewEncryptionBase()
	cryptBase.SetPassphrase(os.Getenv("2FA_KEY_ENCRYPT"))
	encryptSecret, err := cryptBase.Encrypt([]byte(encodedSecret))
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, fmt.Sprintf("error encrypt secret key, err := %s", err.Error()))
		return
	}

	authLink := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s", issuer, account, encodedSecret, issuer)
	code, err := qr.Encode(authLink, qr.H)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	twoAuthModel := models.TwoAuths{
		UserID: currentUser.ID,
		Secret: string(encryptSecret),
		Status: constant.StatusActive,
	}

	err = controller.DB.Create(&twoAuthModel).Error
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

func (controller *TwoFactorAuthController) ValidateAuth(ctx *gin.Context) {
	uuid, _ := ctx.Get("uuid")
	currentUser, _ := helpers.GetCurrentUser(uuid.(string))

	twoAuthModel, err := controller.getUserTwoAuthenticationModel(currentUser.ID)
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

	helpers.NewResponse(ctx, http.StatusOK, fmt.Sprintf("success to authenticate"))
	return
}

func (controller *TwoFactorAuthController) GenerateRecoveryCode(ctx *gin.Context) {
	uuid, _ := ctx.Get("uuid")
	currentUser, _ := helpers.GetCurrentUser(uuid.(string))
	currentUserID := currentUser.ID

	_, err := controller.getUserTwoAuthenticationModel(currentUserID)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusUnauthorized, err.Error())
		return
	}

	tx := controller.DB.Begin()
	var oldRecoveryCodeModels []models.TwoAuthRecoveries
	err = tx.Where(&models.TwoAuthRecoveries{UserID: currentUserID, Status: constant.StatusActive}).Find(&oldRecoveryCodeModels).Error
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	for _, oldRecoveryCodeModel := range oldRecoveryCodeModels {
		oldRecoveryCodeModel.Status = constant.StatusNonActive
		err = tx.Save(&oldRecoveryCodeModel).Error
		if err != nil {
			helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
			return
		}
	}

	var code []string
	for i := 0; i < 10; i++ {
		model := models.TwoAuthRecoveries{
			UserID: currentUserID,
			Code:   crypts.GenerateRandString(8, "number"),
			Status: constant.StatusActive,
		}

		err = tx.Create(&model).Error
		if err != nil {
			tx.Rollback()
			helpers.NewResponse(ctx, http.StatusBadRequest, fmt.Sprintf("error create code, please try again later, err := %s", err.Error()))
			return
		}

		code = append(code, model.Code)
	}

	tx.Commit()
	helpers.NewResponse(ctx, http.StatusUnauthorized, code)
	return
}
