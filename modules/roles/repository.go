package roles

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
	GetDetail(ctx *gin.Context, id string) (*models.Roles, bool, error)
	GetList(ctx *gin.Context, filter RequestListFilter) ([]models.Roles, *paginator.Pagination, error)
	Create(ctx *gin.Context, model models.Roles) (*models.Roles, error)
	Update(ctx *gin.Context, model models.Roles) (*models.Roles, error)
	Delete(ctx *gin.Context, id string) error
}

type repository struct {
	*configs.DI
}

func NewRepository(di *configs.DI) Repository {
	return repository{DI: di}
}

func (r repository) GetDetail(ctx *gin.Context, id string) (*models.Roles, bool, error) {
	db := r.DB

	var model models.Roles
	err := db.Where("id = ?", id).First(&model).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, true, nil
	} else if err != nil {
		return nil, false, err
	}

	return &model, false, nil
}

func (r repository) GetList(ctx *gin.Context, filter RequestListFilter) ([]models.Roles, *paginator.Pagination, error) {
	db := r.DB

	var roleModels []models.Roles
	page, err := paginator.Pages(&paginator.Param{
		DB: db,
		Paging: &paginator.Paging{
			Page:    filter.Page,
			OrderBy: []string{filter.OrderBy},
			Limit:   filter.Limit,
			ShowSQL: false,
		},
	}, &roleModels)
	if err != nil {
		return nil, nil, err
	}

	return roleModels, page, nil
}

func (r repository) Create(ctx *gin.Context, model models.Roles) (*models.Roles, error) {
	db := r.DB
	err := db.Create(&model).Error
	if err != nil {
		return nil, err
	}

	return &model, nil
}

func (r repository) Update(ctx *gin.Context, model models.Roles) (*models.Roles, error) {
	db := r.DB
	err := db.Save(&model).Error
	return &model, err
}

func (r repository) Delete(ctx *gin.Context, id string) error {
	model, notFound, err := r.GetDetail(ctx, id)
	if err != nil {
		return err
	} else if notFound {
		return fmt.Errorf("role %s not found", id)
	}

	db := r.DB
	err = db.Where("id = ?", id).Delete(&model).Error
	return err
}
