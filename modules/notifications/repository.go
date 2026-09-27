package notifications

import (
	"errors"
	"fmt"

	"github.com/dmitryburov/gorm-paginator"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"go-api/constant"
	"go-api/modules/configs"
	"go-api/modules/models"
)

type Repository interface {
	GetDetail(ctx *gin.Context, id int64) (models.NotificationTemplates, bool, error)
	GetList(ctx *gin.Context, filter RequestListFilter) ([]models.NotificationTemplates, *paginator.Pagination, error)
	Create(ctx *gin.Context, model models.NotificationTemplates) (models.NotificationTemplates, error)
	Update(ctx *gin.Context, model models.NotificationTemplates) (models.NotificationTemplates, error)
	Delete(ctx *gin.Context, id int64) error

	QRCodeWhatsapp(ctx *gin.Context) (qrCode string, isExists bool, err error)

	GetUserWithSocials(ctx *gin.Context, userID int64) (models.Users, bool, error)
}

type repository struct {
	*configs.DI
}

func NewRepository(di *configs.DI) Repository {
	return repository{DI: di}
}

func (r repository) QRCodeWhatsapp(ctx *gin.Context) (qrCode string, isExists bool, err error) {
	redisClient := r.RedisSession
	qrCode, err = redisClient.Get(ctx, constant.QRCodeWhatsapp).Result()
	if err == redis.Nil {
		return "", false, nil
	} else if err != nil {
		return "", false, err
	}

	return qrCode, true, nil
}

func (r repository) GetDetail(ctx *gin.Context, id int64) (models.NotificationTemplates, bool, error) {
	db := r.DB.WithContext(ctx)

	var model models.NotificationTemplates
	err := db.First(&model, id).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return models.NotificationTemplates{}, true, nil
	} else if err != nil {
		return models.NotificationTemplates{}, false, err
	}

	return model, false, nil
}

func (r repository) GetList(ctx *gin.Context, filter RequestListFilter) ([]models.NotificationTemplates, *paginator.Pagination, error) {
	db := r.DB.WithContext(ctx)

	if filter.NotificationType != 0 {
		db = db.Where("notification_type = ?", filter.NotificationType)
	}

	if filter.Action != "" {
		db = db.Where("action = ?", filter.Action)
	}

	var templateModels []models.NotificationTemplates
	page, err := paginator.Pages(&paginator.Param{
		DB: db,
		Paging: &paginator.Paging{
			Page:    filter.Page,
			OrderBy: []string{filter.OrderBy},
			Limit:   filter.Limit,
			ShowSQL: false,
		},
	}, &templateModels)
	if err != nil {
		return nil, nil, err
	}

	return templateModels, page, nil
}

func (r repository) Create(ctx *gin.Context, model models.NotificationTemplates) (models.NotificationTemplates, error) {
	db := r.DB.WithContext(ctx)
	err := db.Create(&model).Error
	if err != nil {
		return models.NotificationTemplates{}, err
	}

	return model, nil
}

func (r repository) Update(ctx *gin.Context, model models.NotificationTemplates) (models.NotificationTemplates, error) {
	db := r.DB.WithContext(ctx)
	err := db.Save(&model).Error
	return model, err
}

func (r repository) Delete(ctx *gin.Context, id int64) error {
	_, notFound, err := r.GetDetail(ctx, id)
	if notFound {
		return fmt.Errorf("notification template not found")
	} else if err != nil {
		return err
	}

	db := r.DB.WithContext(ctx)
	return db.Delete(&models.NotificationTemplates{}, id).Error
}

func (r repository) GetUserWithSocials(ctx *gin.Context, userID int64) (models.Users, bool, error) {
	db := r.DB.WithContext(ctx)

	var user models.Users
	err := db.Preload("UserSocials").First(&user, userID).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Users{}, true, nil
	} else if err != nil {
		return models.Users{}, false, err
	}

	return user, false, nil
}
