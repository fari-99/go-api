package twoFA

import (
	"encoding/base32"
	"fmt"
	"os"

	gohelper "github.com/fari-99/go-helper"
	"github.com/fari-99/go-helper/crypts"
	"github.com/gin-gonic/gin"

	"go-api/constant"
	"go-api/helpers"
	"go-api/modules/models"
)

type Service interface {
	GetDetails(ctx *gin.Context, userID uint64) (*models.TwoAuths, bool, error)
	GetUserDetails(ctx *gin.Context, userID uint64) (*models.Users, bool, error)

	// 2FA
	CreateTotp(ctx *gin.Context) (string, string, error)
	UserEnabledTotp(ctx *gin.Context, userID uint64, isEnabled bool) error

	// Recovery Code
	CreateRecoveryCode(ctx *gin.Context, userID uint64) ([]string, error)
	GetAllRecoveryCode(ctx *gin.Context, userID uint64) ([]models.TwoAuthRecoveries, error)
	ValidateRecoveryCode(ctx *gin.Context, recoveryCode string, userID uint64) (bool, error)
	// DeleteAllRecoveryCodes(ctx *gin.Context, userID uint64) error // TODO: create function

	EncryptKey() ([]byte, string, error)
	DecryptKey(twoAuthModel models.TwoAuths) ([]byte, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return service{repo: repo}
}

func (s service) EncryptKey() ([]byte, string, error) {
	secret := gohelper.GenerateRandString(10, "alphanum")
	encodedSecret := base32.StdEncoding.EncodeToString([]byte(secret))

	cryptBase := crypts.NewEncryptionBase()
	cryptBase.SetPassphrase(os.Getenv("2FA_KEY_ENCRYPT"))
	encryptSecret, err := cryptBase.Encrypt([]byte(encodedSecret))
	return encryptSecret, encodedSecret, err
}

func (s service) DecryptKey(twoAuthModel models.TwoAuths) ([]byte, error) {
	cryptBase := crypts.NewEncryptionBase()
	cryptBase.SetPassphrase(os.Getenv("2FA_KEY_ENCRYPT"))
	secret, err := cryptBase.Decrypt([]byte(twoAuthModel.Secret))
	if err != nil {
		return nil, err
	}

	return secret, nil
}

func (s service) GetDetails(ctx *gin.Context, userID uint64) (*models.TwoAuths, bool, error) {
	return s.repo.GetDetails(ctx, userID)
}

func (s service) GetUserDetails(ctx *gin.Context, userID uint64) (*models.Users, bool, error) {
	return s.repo.GetUserDetails(ctx, userID)
}

func (s service) CreateTotp(ctx *gin.Context) (string, string, error) {
	uuid, _ := ctx.Get("uuid")
	currentUser, _ := helpers.GetCurrentUser(ctx, uuid.(string))

	encryptSecret, encodedSecret, err := s.EncryptKey()
	if err != nil {
		return "", "", err
	}

	twoAuthModel := models.TwoAuths{
		UserID:  currentUser.ID,
		Account: currentUser.Email,
		Issuer:  os.Getenv("APP_NAME"),
		Secret:  string(encryptSecret),
		Status:  constant.StatusActive,
	}

	_, err = s.repo.CreateTotp(ctx, twoAuthModel)
	authLink := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s", twoAuthModel.Issuer, twoAuthModel.Account, encodedSecret, twoAuthModel.Issuer)

	return encodedSecret, authLink, err
}

func (s service) UserEnabledTotp(ctx *gin.Context, userID uint64, isEnabled bool) error {
	return s.repo.UserEnabledTotp(ctx, userID, isEnabled)
}

func (s service) CreateRecoveryCode(ctx *gin.Context, userID uint64) ([]string, error) {
	return s.repo.CreateRecoveryCode(ctx, userID)
}

func (s service) GetAllRecoveryCode(ctx *gin.Context, userID uint64) ([]models.TwoAuthRecoveries, error) {
	return s.repo.GetAllRecoveryCode(ctx, userID)
}

func (s service) ValidateRecoveryCode(ctx *gin.Context, recoveryCode string, userID uint64) (bool, error) {
	return s.repo.ValidateRecoveryCode(ctx, recoveryCode, userID)
}
