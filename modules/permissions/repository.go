package permissions

import (
	"go-api/modules/configs"
	"go-api/modules/models"
	"go-api/modules/users"

	"github.com/gin-gonic/gin"
)

// Repository encapsulates the logic to access customers from the data source.
type Repository interface {
	// GetRoles get all roles on this database
	GetRoles(ctx *gin.Context) ([]models.Roles, error)
}

// repository persists customers in database
type repository struct {
	*configs.DI
}

// NewRepository creates a new users repository
func NewRepository(di *configs.DI) Repository {
	return repository{DI: di}
}

func (r repository) GetRoles(ctx *gin.Context) ([]models.Roles, error) {
	userRepo := users.NewRepository(r.DI)
	return userRepo.GetRoles(ctx)
}
