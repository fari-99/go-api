package twoFA

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go-api/constant"
	"go-api/helpers/crypts"
	"go-api/modules/configs"
	"go-api/modules/models"
)

type Repository interface {
	GetDetails(ctx *gin.Context, userID int64) (*models.TwoAuths, bool, error)
	CreateConfigs(ctx *gin.Context, twoAuthModel models.TwoAuths) (models.TwoAuths, error)
	GenerateRecoveryCode(userID int64) ([]string, error)
}

type repository struct {
	*configs.DI
}

func NewRepository(di *configs.DI) Repository {
	return repository{DI: di}
}

func (r repository) GetDetails(ctx *gin.Context, userID int64) (*models.TwoAuths, bool, error) {
	var twoAuthModel models.TwoAuths
	err := r.DB.Where(&models.TwoAuths{UserID: userID, Status: constant.StatusActive}).First(&twoAuthModel).Error
	if err != nil && gorm.IsRecordNotFoundError(err) {
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

func (r repository) GenerateRecoveryCode(userID int64) ([]string, error) {
	tx := r.DB.Begin()
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
			Code:   crypts.GenerateRandString(8, "number"),
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