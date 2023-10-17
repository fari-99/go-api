package storages

import (
	"errors"

	"go-api/modules/configs"
	"go-api/modules/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Repository interface {
	GetDetail(ctx *gin.Context, storageID string) (storageModel *models.Storages, notFound bool, err error)
	Create(ctx *gin.Context, storageModel []models.Storages) ([]models.Storages, error)
}

type repository struct {
	*configs.DI
}

func NewRepository(di *configs.DI) Repository {
	return repository{DI: di}
}

func (r repository) GetDetail(ctx *gin.Context, storageID string) (*models.Storages, bool, error) {
	db := r.DB

	var storageModel models.Storages
	err := db.Where(&models.Storages{Base: models.Base{ID: models.IDType(storageID)}}).First(&storageModel).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, true, nil
	} else if err != nil {
		return nil, false, err
	}

	return &storageModel, false, nil
}

func (r repository) Create(ctx *gin.Context, storageModels []models.Storages) ([]models.Storages, error) {
	tx := r.DB.Begin()

	var savedModels []models.Storages
	for _, storageModel := range storageModels {
		err := tx.Create(&storageModel).Error
		if err != nil {
			tx.Rollback()
			return nil, err
		}

		savedModels = append(savedModels, storageModel)
	}

	tx.Commit()
	return savedModels, nil
}
