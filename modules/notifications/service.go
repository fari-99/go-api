package notifications

import (
	"github.com/biezhi/gorm-paginator/pagination"
	"github.com/gin-gonic/gin"
)

type Service interface {
	GetDetail(ctx *gin.Context, id int64) (interface{}, bool, error)
	GetList(ctx *gin.Context, filter RequestListFilter) ([]interface{}, *pagination.Paginator, error)
	Create(ctx *gin.Context, model interface{}) (interface{}, error)
	Update(ctx *gin.Context, model interface{}) (interface{}, error)
	Delete(ctx *gin.Context, id int64) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return service{repo: repo}
}

func (s service) GetDetail(ctx *gin.Context, id int64) (interface{}, bool, error) {
	return s.repo.GetDetail(ctx, id)
}

func (s service) GetList(ctx *gin.Context, filter RequestListFilter) ([]interface{}, *pagination.Paginator, error) {
	return s.repo.GetList(ctx, filter)
}

func (s service) Create(ctx *gin.Context, model interface{}) (interface{}, error) {
	return s.repo.Create(ctx, model)
}

func (s service) Update(ctx *gin.Context, model interface{}) (interface{}, error) {
	return s.repo.Update(ctx, model)
}

func (s service) Delete(ctx *gin.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
