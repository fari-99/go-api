package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	gohelper "github.com/fari-99/go-helper"
	"github.com/spf13/cast"

	"go-api/constant"
	"go-api/modules/configs"
	"go-api/modules/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Repository interface {
	GetDetails(ctx *gin.Context, userID string) (*models.Users, bool, error)
	GetUserByEmail(ctx *gin.Context, email string) (userModel *models.Users, notFound bool, err error)
	CreateUser(ctx *gin.Context, userModel models.Users) (*models.Users, error)
	GetRoles(ctx *gin.Context) ([]models.Roles, error)
	UpdateUser(ctx *gin.Context, userModel models.Users) (*models.Users, error)
	ForgotPassword(ctx *gin.Context, email string) (userCodes *models.UserCodes, notFound bool, err error)
	ForgotUsername(ctx *gin.Context, email string) (userModel *models.Users, notFound bool, err error)
	ResetPassword(ctx *gin.Context, input ResetPasswordRequest) error
}

type repository struct {
	*configs.DI
}

func NewRepository(di *configs.DI) Repository {
	return repository{DI: di}
}

func (r repository) ResetPassword(ctx *gin.Context, input ResetPasswordRequest) error {
	db := r.DB.WithContext(ctx)

	var userCodeModel models.UserCodes
	err := db.Where(&models.UserCodes{Code: input.Token, IsUsed: constant.UserCodesNew}).
		Where("expired_at > ?", time.Now().Format("2006-01-02 15:04:05")).
		First(&userCodeModel).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("token either already used or expired")
	} else if err != nil {
		return err
	}

	userModel, notFound, err := r.GetDetails(ctx, userCodeModel.UserID)
	if err != nil {
		return err
	} else if notFound {
		return fmt.Errorf("user not found")
	}

	password := gohelper.Passwords{
		Email:    userModel.Email,
		Username: userModel.Username,
		Password: input.Password,
	}

	hashPassword, err := gohelper.GeneratePassword(password, cast.ToInt8(os.Getenv("PASSWORD_COST")))
	if err != nil {
		return err
	}

	userModel.Password = *hashPassword
	_, err = r.UpdateUser(ctx, *userModel)
	return err
}

func (r repository) ForgotUsername(ctx *gin.Context, email string) (userModel *models.Users, notFound bool, err error) {
	userModel, notFound, err = r.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, true, err
	} else if notFound {
		return nil, true, nil
	}

	return userModel, false, nil
}

func (r repository) ForgotPassword(ctx *gin.Context, email string) (userCodes *models.UserCodes, notFound bool, err error) {
	db := r.DB.WithContext(ctx)

	userModel, notFound, err := r.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, true, err
	} else if notFound {
		return nil, true, nil
	}

	var userCodeExists models.UserCodes
	err = db.Where(&models.UserCodes{UserID: userModel.ID, IsUsed: constant.UserCodesNew}).
		Where("expired_at > ?", time.Now().Format("2006-01-02 15:04:05")).First(&userCodeExists).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, true, err
	} else if err == nil {
		return &userCodeExists, false, nil
	}

	params, _ := json.Marshal(map[string]interface{}{"email": email})
	timeNow := time.Now()
	timeExpired := timeNow.AddDate(0, 0, 1)

	userCode := models.UserCodes{
		UserID:    userModel.ID,
		Via:       "email",
		Code:      gohelper.GenerateRandString(10, "alphanum"),
		Params:    string(params),
		IsUsed:    0,
		ExpiredAt: timeExpired,
	}

	err = db.Create(&userCode).Error
	if err != nil {
		return nil, false, err
	}

	return &userCode, false, nil
}

func (r repository) UpdateUser(ctx *gin.Context, userModel models.Users) (*models.Users, error) {
	db := r.DB.WithContext(ctx)
	err := db.Updates(&userModel).Error
	if err != nil {
		return nil, err
	}

	return &userModel, nil
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

func (r repository) GetUserByEmail(ctx *gin.Context, email string) (*models.Users, bool, error) {
	db := r.DB.WithContext(ctx)

	var userModel models.Users
	err := db.Where(&models.Users{Email: email, Status: constant.StatusActive}).First(&userModel).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, true, nil
	} else if err != nil {
		return nil, true, err
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
