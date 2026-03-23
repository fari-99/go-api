package auths

import (
	"encoding/json"
	"errors"

	"github.com/gin-gonic/gin"

	"go-api/helpers"
	"go-api/modules/configs"
	"go-api/modules/models"
	"go-api/modules/users"

	"gorm.io/gorm"
)

type Repository interface {
	AuthenticatePassword(ctx *gin.Context, input RequestAuthUser) (*models.Users, bool, error)
	GetUserDetails(ctx *gin.Context, id uint64) (models.Users, error)
}

type repository struct {
	*configs.DI
}

func NewRepository(di *configs.DI) Repository {
	return repository{DI: di}
}

func (r repository) AuthenticatePassword(ctx *gin.Context, input RequestAuthUser) (*models.Users, bool, error) {
	db := r.DB

	var customerModel models.Users
	err := db.Where(&models.Users{Email: input.Email}).Find(&customerModel).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, true, nil
	}

	err = helpers.PasswordAuth(customerModel.Password, input.Password)
	if err != nil {
		return nil, false, err
	}

	if !customerModel.TwoFaEnabled {
		return &customerModel, false, nil
	}

	userModel, err := r.GetUserDetails(ctx, customerModel.ID.Uint64())
	return &userModel, false, err
}

func (r repository) GetUserDetails(ctx *gin.Context, id uint64) (models.Users, error) {
	userService := users.NewService(users.NewRepository(r.DI))
	userProfile, err := userService.UserProfile(ctx, id)

	var userModel models.Users
	userProfileMarshal, _ := json.Marshal(userProfile)
	_ = json.Unmarshal(userProfileMarshal, &userModel)

	return userModel, err
}
