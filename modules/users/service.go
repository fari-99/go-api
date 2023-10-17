package users

import (
	"fmt"
	"os"

	gohelper "github.com/fari-99/go-helper"
	"github.com/spf13/cast"

	"go-api/constant"
	"go-api/helpers"
	"go-api/helpers/notifications"
	"go-api/modules/models"

	"github.com/gin-gonic/gin"
)

type Service interface {
	CreateUser(ctx *gin.Context, input RequestCreateUser) (*models.Users, error)
	UserProfile(ctx *gin.Context, userID string) (models.UserProfile, error)
	ChangePassword(ctx *gin.Context, input RequestChangePassword) (exists bool, err error)
	ForgotPassword(ctx *gin.Context, input ForgotPasswordRequest) (userCodes *models.UserCodes, notFound bool, err error)
	ForgotUsername(ctx *gin.Context, input ForgotUsernameRequest) (exists bool, err error)
	ResetPassword(ctx *gin.Context, input ResetPasswordRequest) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return service{repo: repo}
}

func (s service) ResetPassword(ctx *gin.Context, input ResetPasswordRequest) error {
	if err := input.Validate(); err != nil {
		return err
	}

	err := s.repo.ResetPassword(ctx, input)
	return err
}

func (s service) ForgotPassword(ctx *gin.Context, input ForgotPasswordRequest) (userCodes *models.UserCodes, notFound bool, err error) {
	if err = input.Validate(); err != nil {
		return nil, false, err
	}

	userCode, notFound, err := s.repo.ForgotPassword(ctx, input.Email)
	if err != nil {
		return nil, false, err
	} else if notFound {
		return nil, notFound, nil
	}

	return userCode, false, nil
}

func (s service) ForgotUsername(ctx *gin.Context, input ForgotUsernameRequest) (exists bool, err error) {
	if err = input.Validate(); err != nil {
		return false, err
	}

	userModel, notFound, err := s.repo.ForgotUsername(ctx, input.Email)
	if err != nil {
		return false, err
	} else if notFound {
		return false, nil
	}

	// send emails
	emails := notifications.Email{
		Subject: "Your forgotten username",
		Body:    fmt.Sprintf("your username is <strong>%s</strong>", userModel.Username),
		From:    "no-reply@fadhlan.com",
		To:      []string{input.Email},
	}

	err = notifications.SendEmail(emails)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s service) ChangePassword(ctx *gin.Context, input RequestChangePassword) (exists bool, err error) {
	uuidSession, _ := ctx.Get("uuid")
	currentUser, _ := helpers.GetCurrentUser(uuidSession.(string))

	if err = input.Validate(); err != nil {
		return false, err
	}

	userModel, notFound, err := s.repo.GetDetails(ctx, string(currentUser.ID))
	if err != nil {
		return false, err
	} else if notFound {
		return notFound, nil
	}

	err = helpers.PasswordAuth(userModel.Password, input.CurrentPassword)
	if err != nil { // Password not match!!
		return true, err
	}

	password := gohelper.Passwords{
		Email:    userModel.Email,
		Username: userModel.Username,
		Password: input.NewPassword,
	}

	hashPassword, err := gohelper.GeneratePassword(password, cast.ToInt8(os.Getenv("PASSWORD_COST")))
	if err != nil {
		return true, err
	}

	userModel.Password = *hashPassword
	_, err = s.repo.UpdateUser(ctx, *userModel)
	return true, err
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
