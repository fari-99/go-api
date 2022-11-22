package users

import (
	"fmt"
	"os"

	gohelper "github.com/fari-99/go-helper"
	"github.com/spf13/cast"

	"go-api/constant"
	"go-api/modules/models"

	"github.com/gin-gonic/gin"
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

	password := gohelper.Passwords{
		Email:    input.Email,
		Username: input.Username,
		Password: input.Password,
	}

	hashPassword, err := gohelper.GeneratePassword(password, cast.ToInt8(os.Getenv("PASSWORD_COST")))
	if err != nil {
		return nil, err
	}

	userModel := models.Users{
		Username: input.Username,
		Password: *hashPassword,
		Email:    input.Email,
		Status:   constant.StatusActive,
	}

	savedModel, err := s.repo.CreateUser(ctx, userModel)
	return savedModel, err
}
