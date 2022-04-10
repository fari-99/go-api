package users

import (
	"fmt"
	"go-api/constant"
	"go-api/helpers"
	"go-api/modules/models"

	"github.com/gin-gonic/gin"
	"github.com/nbutton23/zxcvbn-go"
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

	var userInput []string
	checkPassword := zxcvbn.PasswordStrength(input.Password, userInput)
	if checkPassword.Score <= 2 {
		return nil, fmt.Errorf("your password not good enough, please try again")
	}

	password, err := helpers.GeneratePassword(input.Password)
	if err != nil {
		return nil, err
	}

	userModel := models.Users{
		Username: input.Username,
		Password: password,
		Email:    input.Email,
		Status:   constant.StatusActive,
	}

	savedModel, err := s.repo.CreateUser(ctx, userModel)
	return savedModel, err
}
