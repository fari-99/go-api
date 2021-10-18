package users

import (
	"github.com/gin-gonic/gin"
	"go-api/constant"
	"go-api/helpers"
	"go-api/modules/models"
)

type Service interface {
	CreateUser(ctx *gin.Context, input RequestCreateUser) (*models.Users, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return service{repo: repo}
}

func (s service) CreateUser(ctx *gin.Context, input RequestCreateUser) (*models.Users, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	password, err := helpers.GeneratePassword(input.Password)
	if err != nil {
		return nil, err
	}

	userModel := models.Users{
		Username:  input.Username,
		Password:  password,
		Email:     input.Email,
		Status:    constant.StatusActive,
	}

	savedModel, err := s.repo.CreateUser(ctx,userModel)
	return savedModel, err
}