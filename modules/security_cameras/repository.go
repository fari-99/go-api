package security_cameras

import (
	"errors"
	"fmt"

	"go-api/modules/configs"
	"go-api/modules/models"

	paginator "github.com/dmitryburov/gorm-paginator"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Repository interface {
	GetDetail(ctx *gin.Context, id int64) (*models.SecurityCameras, bool, error)
	GetList(ctx *gin.Context, filter RequestListFilter) ([]models.SecurityCameras, *paginator.Pagination, error)
	Create(ctx *gin.Context, model models.SecurityCameras) (*models.SecurityCameras, error)
	Update(ctx *gin.Context, model models.SecurityCameras) (*models.SecurityCameras, error)
	Delete(ctx *gin.Context, id int64) error
}

type repository struct {
	*configs.DI
}

func NewRepository(di *configs.DI) Repository {
	return repository{DI: di}
}

func (r repository) GetDetail(ctx *gin.Context, id int64) (*models.SecurityCameras, bool, error) {
	db := r.DB

	var model models.SecurityCameras
	err := db.First(&model, id).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, true, nil
	} else if err != nil {
		return nil, false, err
	}

	return &model, false, nil
}

func (r repository) GetList(ctx *gin.Context, filter RequestListFilter) ([]models.SecurityCameras, *paginator.Pagination, error) {
	db := r.DB

	var secCameraModels []models.SecurityCameras
	page, err := paginator.Pages(&paginator.Param{
		DB: db,
		Paging: &paginator.Paging{
			Page:    filter.Page,
			OrderBy: []string{filter.OrderBy},
			Limit:   filter.Limit,
			ShowSQL: false,
		},
	}, &secCameraModels)
	if err != nil {
		return nil, nil, err
	}

	return secCameraModels, page, nil
}

func (r repository) Create(ctx *gin.Context, model models.SecurityCameras) (*models.SecurityCameras, error) {
	db := r.DB
	err := db.Create(&model).Error
	if err != nil {
		return nil, err
	}

	return &model, nil
}

func (r repository) Update(ctx *gin.Context, model models.SecurityCameras) (*models.SecurityCameras, error) {
	db := r.DB
	err := db.Save(&model).Error
	return &model, err
}

func (r repository) Delete(ctx *gin.Context, id int64) error {
	model, notFound, err := r.GetDetail(ctx, id)
	if notFound {
		return fmt.Errorf("model not found")
	} else if err != nil {
		return err
	}

	db := r.DB
	err = db.Delete(&model, id).Error
	return err
}
