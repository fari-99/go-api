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
	GetDetails(ctx *gin.Context, userID string) (*models.TwoAuths, bool, error)
	CreateConfigs(ctx *gin.Context) (models.TwoAuths, string, error)
	GenerateRecoveryCode(ctx *gin.Context, userID string) ([]string, error)
	GetAllRecoveryCode(ctx *gin.Context, userID string) ([]models.TwoAuthRecoveries, error)

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

func (s service) GetDetails(ctx *gin.Context, userID string) (*models.TwoAuths, bool, error) {
	return s.repo.GetDetails(ctx, userID)
}

func (s service) CreateConfigs(ctx *gin.Context) (models.TwoAuths, string, error) {
	uuid, _ := ctx.Get("uuid")
	currentUser, _ := helpers.GetCurrentUser(uuid.(string))

	encryptSecret, encodedSecret, err := s.EncryptKey()
	if err != nil {
		return models.TwoAuths{}, "", err
	}

	twoAuthModel := models.TwoAuths{
		UserID:  currentUser.ID,
		Account: currentUser.Email,
		Issuer:  os.Getenv("APP_NAME"),
		Secret:  string(encryptSecret),
		Status:  constant.StatusActive,
	}

	savedModel, err := s.repo.CreateConfigs(ctx, twoAuthModel)
	authLink := fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s", twoAuthModel.Issuer, twoAuthModel.Account, encodedSecret, twoAuthModel.Issuer)

	return savedModel, authLink, err
}

func (s service) GenerateRecoveryCode(ctx *gin.Context, userID string) ([]string, error) {
	return s.repo.GenerateRecoveryCode(ctx, userID)
}

func (s service) GetAllRecoveryCode(ctx *gin.Context, userID string) ([]models.TwoAuthRecoveries, error) {
	return s.repo.GetAllRecoveryCode(ctx, userID)
}
