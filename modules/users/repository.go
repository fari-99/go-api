package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	gohelper "github.com/fari-99/go-helper"
	"github.com/spf13/cast"

	"go-api/constant"
	"go-api/constant/constant_models"
	"go-api/modules/configs"
	"go-api/modules/models"

	paginator "github.com/dmitryburov/gorm-paginator"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Repository interface {
	GetDetails(ctx *gin.Context, userID uint64) (*models.Users, bool, error)
	GetUserByEmail(ctx *gin.Context, email string) (userModel *models.Users, notFound bool, err error)
	CreateUser(ctx *gin.Context, userModel models.Users) (*models.Users, error)
	GetRoles(ctx *gin.Context) ([]models.Roles, error)
	GetUserRoleIDs(ctx *gin.Context, userID uint64) ([]uint64, error)
	SetUserRoles(ctx *gin.Context, userID uint64, roleIDs []uint64) error
	GetList(ctx *gin.Context, filter RequestListUsers) ([]models.Users, *paginator.Pagination, error)
	UpdateUser(ctx *gin.Context, userModel models.Users) (*models.Users, error)
	DeleteUser(ctx *gin.Context, userID uint64) error
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

	userModel, notFound, err := r.GetDetails(ctx, userCodeModel.UserID.Uint64())
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

func (r repository) GetUserRoleIDs(ctx *gin.Context, userID uint64) ([]uint64, error) {
	db := r.DB.WithContext(ctx)

	var userRoles []models.UserRoles
	err := db.Where(&models.UserRoles{UserID: models.IDType(userID)}).Find(&userRoles).Error
	if err != nil {
		return nil, err
	}

	roleIDs := make([]uint64, 0, len(userRoles))
	for _, userRole := range userRoles {
		roleIDs = append(roleIDs, userRole.RoleID.Uint64())
	}

	return roleIDs, nil
}

func (r repository) SetUserRoles(ctx *gin.Context, userID uint64, roleIDs []uint64) error {
	db := r.DB.WithContext(ctx)

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where(&models.UserRoles{UserID: models.IDType(userID)}).Delete(&models.UserRoles{}).Error; err != nil {
			return err
		}

		for _, roleID := range roleIDs {
			userRole := models.UserRoles{UserID: models.IDType(userID), RoleID: models.IDType(roleID)}
			if err := tx.Create(&userRole).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// populateUserRoles fills userModel.Roles with the comma-separated Casbin
// subjects ("{RoleName}-{UserType}") for every role assigned via user_roles,
// matching the subject format the permissions/roles modules use for policies.
func (r repository) populateUserRoles(db *gorm.DB, userModel *models.Users) error {
	var userRoles []models.UserRoles
	if err := db.Where(&models.UserRoles{UserID: userModel.ID}).Find(&userRoles).Error; err != nil {
		return err
	} else if len(userRoles) == 0 {
		return nil
	}

	roleIDs := make([]models.IDType, 0, len(userRoles))
	for _, userRole := range userRoles {
		roleIDs = append(roleIDs, userRole.RoleID)
	}

	var roleModels []models.Roles
	if err := db.Where("id IN ?", roleIDs).Find(&roleModels).Error; err != nil {
		return err
	}

	userTypes := constant_models.GetUserTypes()
	subjects := make([]string, 0, len(roleModels))
	for _, role := range roleModels {
		subjects = append(subjects, fmt.Sprintf("%s-%s", role.RoleName, userTypes[int(role.RoleType)]))
	}

	userModel.Roles = strings.Join(subjects, ",")
	return nil
}

func (r repository) GetList(ctx *gin.Context, filter RequestListUsers) ([]models.Users, *paginator.Pagination, error) {
	db := r.DB.WithContext(ctx)
	if filter.Search != "" {
		search := "%" + filter.Search + "%"
		db = db.Where("username LIKE ? OR email LIKE ?", search, search)
	}

	var userModels []models.Users
	page, err := paginator.Pages(&paginator.Param{
		DB: db,
		Paging: &paginator.Paging{
			Page:    filter.Page,
			OrderBy: []string{filter.OrderBy},
			Limit:   filter.Limit,
			ShowSQL: false,
		},
	}, &userModels)
	if err != nil {
		return nil, nil, err
	}

	return userModels, page, nil
}

func (r repository) DeleteUser(ctx *gin.Context, userID uint64) error {
	db := r.DB.WithContext(ctx)
	return db.Where("id = ?", userID).Delete(&models.Users{}).Error
}

func (r repository) GetDetails(ctx *gin.Context, userID uint64) (*models.Users, bool, error) {
	db := r.DB.WithContext(ctx)

	var userModel models.Users
	err := db.Where(&models.Users{Base: models.Base{ID: models.IDType(userID)}}).First(&userModel).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, true, nil
	} else if err != nil {
		return nil, false, err
	}

	if err = r.populateUserRoles(db, &userModel); err != nil {
		return nil, false, err
	}

	if !userModel.TwoFaEnabled {
		return &userModel, false, nil
	}

	userModel.TwoFaModels = &models.TwoAuthsModels{}

	var twoFaModel models.TwoAuths
	err = db.Where(&models.TwoAuths{UserID: userModel.ID}).First(&twoFaModel).Error
	if err == nil {
		userModel.TwoFaModels.TOTP = true
	}

	var recoveryCodeModel models.TwoAuthRecoveries
	err = db.Where(&models.TwoAuthRecoveries{UserID: userModel.ID}).First(&recoveryCodeModel).Error
	if err == nil {
		userModel.TwoFaModels.RecoveryCode = true
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
