package notifications

import (
	"fmt"

	"github.com/dmitryburov/gorm-paginator"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"go-api/constant"
	"go-api/modules/configs"
)

type Repository interface {
	GetDetail(ctx *gin.Context, id int64) (interface{}, bool, error)
	GetList(ctx *gin.Context, filter RequestListFilter) ([]interface{}, *paginator.Pagination, error)
	Create(ctx *gin.Context, model interface{}) (interface{}, error)
	Update(ctx *gin.Context, model interface{}) (interface{}, error)
	Delete(ctx *gin.Context, id int64) error

	QRCodeWhatsapp(ctx *gin.Context) (qrCode string, isExists bool, err error)
}

type repository struct {
	*configs.DI
}

func NewRepository(di *configs.DI) Repository {
	return repository{DI: di}
}

func (r repository) QRCodeWhatsapp(ctx *gin.Context) (qrCode string, isExists bool, err error) {
	redisClient := r.Redis
	qrCode, err = redisClient.Get(ctx, constant.QRCodeWhatsapp).Result()
	if err == redis.Nil {
		return "", false, nil
	} else if err != nil {
		return "", false, err
	}

	return qrCode, true, nil
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

func (r repository) GetList(ctx *gin.Context, filter RequestListFilter) ([]interface{}, *paginator.Pagination, error) {
	db := r.DB

	var models []interface{}
	page, err := paginator.Pages(&paginator.Param{
		DB: db,
		Paging: &paginator.Paging{
			Page:    filter.Page,
			OrderBy: []string{filter.OrderBy},
			Limit:   filter.Limit,
			ShowSQL: false,
		},
	}, &models)
	if err != nil {
		return nil, nil, err
	}

	return models, page, nil
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
