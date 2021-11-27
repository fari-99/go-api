package permissions

import (
	"go-api/modules/models"

	"github.com/gin-gonic/gin"
)

// Service encapsulates usecase logic for customers.
type Service interface {
	GetRoles(ctx *gin.Context) ([]models.Roles, error)
}

type service struct {
	repo Repository
}

// NewService creates a new users service.
func NewService(repo Repository) Service {
	return service{repo}
}

func (s service) GetRoles(ctx *gin.Context) ([]models.Roles, error) {
	return s.repo.GetRoles(ctx)
}
