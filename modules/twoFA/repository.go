package twoFA

import (
	"errors"

	gohelper "github.com/fari-99/go-helper"

	"go-api/constant"
	"go-api/modules/configs"
	"go-api/modules/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Repository interface {
	GetDetails(ctx *gin.Context, userID string) (*models.TwoAuths, bool, error)
	CreateConfigs(ctx *gin.Context, twoAuthModel models.TwoAuths) (models.TwoAuths, error)
	GenerateRecoveryCode(ctx *gin.Context, userID string) ([]string, error)
	GetAllRecoveryCode(ctx *gin.Context, userID string) (recoveryCodeModels []models.TwoAuthRecoveries, err error)
}

type repository struct {
	*configs.DI
}

func NewRepository(di *configs.DI) Repository {
	return repository{DI: di}
}

func (r repository) GetDetails(ctx *gin.Context, userID string) (*models.TwoAuths, bool, error) {
	var twoAuthModel models.TwoAuths
	err := r.DB.Where(&models.TwoAuths{UserID: userID, Status: constant.StatusActive}).First(&twoAuthModel).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, true, nil
	} else if err != nil {
		return nil, false, err
	}

	return &twoAuthModel, false, nil
}

func (r repository) CreateConfigs(ctx *gin.Context, twoAuthModel models.TwoAuths) (models.TwoAuths, error) {
	err := r.DB.Create(&twoAuthModel).Error
	if err != nil {
		return models.TwoAuths{}, err
	}

	return twoAuthModel, nil
}

func (r repository) GetAllRecoveryCode(ctx *gin.Context, userID string) ([]models.TwoAuthRecoveries, error) {
	db := r.DB.WithContext(ctx)

	var recoveryCodes []models.TwoAuthRecoveries
	err := db.Where(&models.TwoAuthRecoveries{UserID: userID}).Find(&recoveryCodes).Error
	if err != nil {
		return nil, err
	}

	return recoveryCodes, nil
}

func (r repository) GenerateRecoveryCode(ctx *gin.Context, userID string) ([]string, error) {
	db := r.DB.WithContext(ctx)
	tx := db.Begin()
	var oldRecoveryCodeModels []models.TwoAuthRecoveries
	err := tx.Where(&models.TwoAuthRecoveries{UserID: userID, Status: constant.StatusActive}).Find(&oldRecoveryCodeModels).Error
	if err != nil {
		return nil, err
	}

	for _, oldRecoveryCodeModel := range oldRecoveryCodeModels {
		oldRecoveryCodeModel.Status = constant.StatusNonActive
		err = tx.Save(&oldRecoveryCodeModel).Error
		if err != nil {
			return nil, err
		}
	}

	var code []string
	for i := 0; i < 10; i++ {
		model := models.TwoAuthRecoveries{
			UserID: userID,
			Code:   gohelper.GenerateRandString(8, "number"),
			Status: constant.StatusActive,
		}

		err = tx.Create(&model).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		code = append(code, model.Code)
	}

	tx.Commit()
	return code, nil
}
