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
	UserProfile(ctx *gin.Context, userID string) (models.UserProfile, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return service{repo: repo}
}

func (s service) UserProfile(ctx *gin.Context, userID string) (models.UserProfile, error) {
	userModel, notFound, err := s.repo.GetDetails(ctx, userID)
	if err != nil {
		return models.UserProfile{}, err
	} else if notFound {
		return models.UserProfile{}, fmt.Errorf("user not found")
	}

	userProfile := models.UserProfile{
		Username:  userModel.Username,
		Email:     userModel.Email,
		Status:    userModel.Status,
		CreatedAt: userModel.CreatedAt,
		UpdatedAt: userModel.UpdatedAt,
	}

	return userProfile, nil
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
