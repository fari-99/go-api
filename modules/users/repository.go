package users

import (
	"errors"
	"fmt"

	"go-api/modules/configs"
	"go-api/modules/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Repository interface {
	GetDetails(ctx *gin.Context, userID string) (*models.Users, bool, error)
	CreateUser(ctx *gin.Context, userModel models.Users) (*models.Users, error)
	GetRoles(ctx *gin.Context) ([]models.Roles, error)
}

type repository struct {
	*configs.DI
}

func NewRepository(di *configs.DI) Repository {
	return repository{DI: di}
}

func (r repository) GetRoles(ctx *gin.Context) ([]models.Roles, error) {
	var roles []models.Roles
	err := r.DB.WithContext(ctx).Find(&roles).Error
	if err != nil {
		return nil, err
	}

	return roles, nil
}

func (r repository) GetDetails(ctx *gin.Context, userID string) (*models.Users, bool, error) {
	db := r.DB.WithContext(ctx)

	var userModel models.Users
	err := db.Where(&models.Users{Base: models.Base{ID: userID}}).First(&userModel).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, true, nil
	} else if err != nil {
		return nil, false, err
	}

	return &userModel, false, nil
}

func (r repository) CreateUser(ctx *gin.Context, userModel models.Users) (*models.Users, error) {
	db := r.DB.WithContext(ctx)

	var isExist models.Users
	err := db.Debug().Where("username = ? OR email = ?", userModel.Username, userModel.Email).Find(&isExist).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("user with that username or email already created")
	}

	err = db.Create(&userModel).Error
	if err != nil {
		return nil, err
	}

	return &userModel, nil
}
