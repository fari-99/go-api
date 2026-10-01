package storages

import (
	"errors"

	"go-api/constant"
	"go-api/modules/configs"
	"go-api/modules/models"

	paginator "github.com/dmitryburov/gorm-paginator"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Repository interface {
	GetDetail(ctx *gin.Context, storageID uint64) (storageModel *models.Storages, notFound bool, err error)
	Create(ctx *gin.Context, storageModel []models.Storages) ([]models.Storages, error)
	GetList(ctx *gin.Context, createdBy uint64, page, limit int) ([]models.Storages, *paginator.Pagination, error)
}

type repository struct {
	*configs.DI
}

func NewRepository(di *configs.DI) Repository {
	return repository{DI: di}
}

func (r repository) GetDetail(ctx *gin.Context, storageID uint64) (*models.Storages, bool, error) {
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

// GetList returns the active storages created by the user, newest first.
func (r repository) GetList(ctx *gin.Context, createdBy uint64, page, limit int) ([]models.Storages, *paginator.Pagination, error) {
	db := r.DB.Where("created_by = ? AND status = ?", createdBy, constant.StatusActive)

	var storageModels []models.Storages
	result, err := paginator.Pages(&paginator.Param{
		DB: db,
		Paging: &paginator.Paging{
			Page:    page,
			Limit:   limit,
			OrderBy: []string{"id desc"},
		},
	}, &storageModels)
	if err != nil {
		return nil, nil, err
	}

	return storageModels, result, nil
}
