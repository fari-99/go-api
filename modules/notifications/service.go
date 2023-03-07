package notifications

import (
	paginator "github.com/dmitryburov/gorm-paginator"
	"github.com/gin-gonic/gin"
)

type Service interface {
	GetDetail(ctx *gin.Context, id int64) (interface{}, bool, error)
	GetList(ctx *gin.Context, filter RequestListFilter) ([]interface{}, *paginator.Pagination, error)
	Create(ctx *gin.Context, model interface{}) (interface{}, error)
	Update(ctx *gin.Context, model interface{}) (interface{}, error)
	Delete(ctx *gin.Context, id int64) error

	QRCodeWhatsapp(ctx *gin.Context) (qrCode string, isExists bool, err error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return service{repo: repo}
}

func (s service) QRCodeWhatsapp(ctx *gin.Context) (qrCode string, isExists bool, err error) {
	return s.repo.QRCodeWhatsapp(ctx)
}

func (s service) GetDetail(ctx *gin.Context, id int64) (interface{}, bool, error) {
	return s.repo.GetDetail(ctx, id)
}

func (s service) GetList(ctx *gin.Context, filter RequestListFilter) ([]interface{}, *paginator.Pagination, error) {
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
