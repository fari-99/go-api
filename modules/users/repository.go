package users

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go-api/modules/configs"
	"go-api/modules/models"
)

type Repository interface {
	GetDetails(ctx *gin.Context, userID int64) (*models.Users, bool, error)
	CreateUser(ctx *gin.Context, userModel models.Users) (*models.Users, error)
}

type repository struct {
	*configs.DI
}

func NewRepository(di *configs.DI) Repository {
	return repository{DI: di}
}

func (r repository) GetDetails(ctx *gin.Context, userID int64) (*models.Users, bool, error) {
	db := r.DB

	var userModel models.Users
	err := db.Where(&models.Users{ID: userID}).First(&userModel).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, true, nil
	} else if err != nil {
		return nil, false, err
	}

	return &userModel, false, nil
}

func (r repository) CreateUser(ctx *gin.Context, userModel models.Users) (*models.Users, error) {
	db := r.DB

	var isExist models.Users
	if !db.Debug().Where("username = ? OR email = ?", userModel.Username, userModel.Email).Find(&isExist).RecordNotFound() {
		return nil, fmt.Errorf("user with that username or email already created")
	}

	err := db.Create(&userModel).Error
	if err != nil {
		return nil, err
	}

	return &userModel, nil
}