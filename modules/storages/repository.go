package storages

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go-api/modules/configs"
	"go-api/modules/models"
)

type Repository interface {
	GetDetail(ctx *gin.Context, storageID int64) (storageModel *models.Storages, notFound bool, err error)
	Create(ctx *gin.Context, storageModel models.Storages) (models.Storages, error)
}

type repository struct {
	*configs.DI
}

func NewRepository(di *configs.DI) Repository {
	return repository{DI: di}
}

func (r repository) GetDetail(ctx *gin.Context, storageID int64) (*models.Storages, bool, error) {
	db := r.DB

	var storageModel models.Storages
	err := db.Where(&models.Storages{Id: storageID}).First(&storageModel).Error
	if err != nil && gorm.IsRecordNotFoundError(err) {
		return nil, true, nil
	} else if err != nil {
		return nil, false, err
	}

	return &storageModel, false, nil
}

func (r repository) Create(ctx *gin.Context, storageModel models.Storages) (models.Storages, error) {
	db := r.DB
	err := db.Create(&storageModel).Error
	if err != nil {
		return models.Storages{}, err
	}

	return storageModel, nil
}