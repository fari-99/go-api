package copy_modules

import (
	"fmt"
	"github.com/biezhi/gorm-paginator/pagination"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go-api/modules/configs"
)

type Repository interface {
	GetDetail(ctx *gin.Context, id int64) (interface{}, bool, error)
	GetList(ctx *gin.Context, filter RequestListFilter) ([]interface{}, *pagination.Paginator, error)
	Create(ctx *gin.Context, model interface{}) (interface{}, error)
	Update(ctx *gin.Context, model interface{}) (interface{}, error)
	Delete(ctx *gin.Context, id int64) error
}

type repository struct {
	*configs.DI
}

func NewRepository(di *configs.DI) Repository {
	return repository{DI: di}
}

func (r repository) GetDetail(ctx *gin.Context, id int64) (interface{}, bool, error) {
	db := r.DB

	var model interface{}
	err := db.First(&model, id).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return nil, true, nil
	} else if err != nil {
		return nil, false, err
	}

	return model, false, nil
}

func (r repository) GetList(ctx *gin.Context, filter RequestListFilter) ([]interface{}, *pagination.Paginator, error) {
	db := r.DB

	var models []interface{}
	paginator := pagination.Paging(&pagination.Param{
		DB:      db,
		Page:    filter.Page,
		Limit:   filter.Limit,
		OrderBy: []string{filter.OrderBy},
		ShowSQL: false,
	}, &models)

	return models, paginator, nil
}

func (r repository) Create(ctx *gin.Context, model interface{}) (interface{}, error) {
	db := r.DB
	err := db.Create(&model).Error
	if err != nil {
		return nil, err
	}

	return model, nil
}

func (r repository) Update(ctx *gin.Context, model interface{}) (interface{}, error) {
	db := r.DB
	err := db.Save(&model).Error
	return model, err
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
