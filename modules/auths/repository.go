package auths

import (
	"go-api/helpers"
	"go-api/modules/configs"
	"go-api/modules/models"
)

type Repository interface {
	AuthenticatePassword(input RequestAuthUser) (*models.Users, bool, error)
}

type repository struct {
	*configs.DI
}

func NewRepository(di *configs.DI) Repository {
	return repository{DI: di}
}

func (r repository) AuthenticatePassword(input RequestAuthUser) (*models.Users, bool, error) {
	db := r.DB

	var customerModel models.Users
	if db.Where(&models.Users{Email: input.Email}).Find(&customerModel).RecordNotFound() {
		return nil, true, nil
	}

	err := helpers.AuthenticatePassword(&customerModel, input.Password)
	if err != nil {
		return nil, false, err
	}

	return &customerModel, false, nil
}
