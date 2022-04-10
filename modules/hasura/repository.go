package hasura

import (
	"go-api/modules/configs"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Repository interface {
	GetDetail(ctx *gin.Context, id int64) (interface{}, bool, error)
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
