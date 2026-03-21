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
	GetDetails(ctx *gin.Context, userID uint64) (*models.TwoAuths, bool, error)

	// 2FA
	CreateConfigs(ctx *gin.Context, twoAuthModel models.TwoAuths) (models.TwoAuths, error)
	TwoFAUserUpdate(ctx *gin.Context, userID uint64, isEnabled bool) error

	// Recovery Code
	GenerateRecoveryCode(ctx *gin.Context, userID uint64) ([]string, error)
	GetAllRecoveryCode(ctx *gin.Context, userID uint64) (recoveryCodeModels []models.TwoAuthRecoveries, err error)
	ValidateRecoveryCode(ctx *gin.Context, recoveryCode string, userID uint64) (bool, error)
}

type repository struct {
	*configs.DI
}

func NewRepository(di *configs.DI) Repository {
	return repository{DI: di}
}

func (r repository) GetDetails(ctx *gin.Context, userID uint64) (*models.TwoAuths, bool, error) {
	var twoAuthModel models.TwoAuths
	err := r.DB.Where(&models.TwoAuths{UserID: models.IDType(userID), Status: constant.StatusActive}).First(&twoAuthModel).Error
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

func (r repository) TwoFAUserUpdate(ctx *gin.Context, userID uint64, isEnabled bool) error {
	db := r.DB.WithContext(ctx)

	dbTx := db.Begin()
	var errDB error
	defer func() {
		if r := recover(); r != nil {
			dbTx.Rollback()
			return
		}

		if errDB != nil {
			dbTx.Rollback()
			return
		}
	}()

	errDB = dbTx.Table("users").
		Where("id = ?", userID).
		Update("two_fa_enabled", isEnabled).Error
	if errDB != nil {
		return errDB
	}

	if !isEnabled {
		var twoFaModel models.TwoAuths
		errDB = dbTx.Where("user_id = ?", userID).First(&twoFaModel).Error
		if errDB != nil {
			return errDB
		}

		errDB = dbTx.Delete(&twoFaModel).Error
		if errDB != nil {
			return errDB
		}
	}

	dbTx.Commit()
	return nil
}

func (r repository) GetAllRecoveryCode(ctx *gin.Context, userID uint64) ([]models.TwoAuthRecoveries, error) {
	db := r.DB.WithContext(ctx)

	var recoveryCodes []models.TwoAuthRecoveries
	err := db.Where(&models.TwoAuthRecoveries{UserID: models.IDType(userID)}).Find(&recoveryCodes).Error
	if err != nil {
		return nil, err
	}

	return recoveryCodes, nil
}

func (r repository) GenerateRecoveryCode(ctx *gin.Context, userID uint64) ([]string, error) {
	db := r.DB.WithContext(ctx)
	tx := db.Begin()
	var oldRecoveryCodeModels []models.TwoAuthRecoveries
	err := tx.Where(&models.TwoAuthRecoveries{UserID: models.IDType(userID), Status: constant.StatusActive}).Find(&oldRecoveryCodeModels).Error
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
			UserID: models.IDType(userID),
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

func (r repository) ValidateRecoveryCode(ctx *gin.Context, recoveryCode string, userID uint64) (bool, error) {
	db := r.DB.WithContext(ctx)

	var recoveryCodeModel models.TwoAuthRecoveries
	err := db.Where(&models.TwoAuthRecoveries{
		UserID: models.IDType(userID),
		Code:   recoveryCode,
	}).First(&recoveryCodeModel).Error
	if err != nil {
		return false, err
	}

	err = db.Delete(&recoveryCodeModel).Error
	if err != nil {
		return false, err
	}

	return true, nil
}
