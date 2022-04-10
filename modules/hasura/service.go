package hasura

import (
	"github.com/gin-gonic/gin"
)

type Service interface {
	GetDetail(ctx *gin.Context, id int64) (interface{}, bool, error)
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
