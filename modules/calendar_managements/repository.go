package calendar_managements

import (
	"fmt"

	"go-api/modules/configs"
	"go-api/modules/models"

	paginator "github.com/dmitryburov/gorm-paginator"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Repository interface {
	GetDetail(ctx *gin.Context, id string) (models.CalendarManagements, bool, error)
	GetList(ctx *gin.Context, filter RequestListFilter) ([]models.CalendarManagements, *paginator.Pagination, error)
	Create(ctx *gin.Context, calendarManagement models.CalendarManagements) (models.CalendarManagements, error)
	Update(ctx *gin.Context, calendarManagement models.CalendarManagements) (models.CalendarManagements, error)
	Delete(ctx *gin.Context, id string) error
}

type repository struct {
	*configs.DI
}

func NewRepository(di *configs.DI) Repository {
	return repository{DI: di}
}

func (r repository) GetDetail(ctx *gin.Context, id string) (models.CalendarManagements, bool, error) {
	db := r.DB

	var calendarManagement models.CalendarManagements
	err := db.First(&calendarManagement, id).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return models.CalendarManagements{}, true, nil
	} else if err != nil {
		return models.CalendarManagements{}, false, err
	}

	return calendarManagement, false, nil
}

func (r repository) GetList(ctx *gin.Context, filter RequestListFilter) ([]models.CalendarManagements, *paginator.Pagination, error) {
	db := r.DB

	var calendarManagements []models.CalendarManagements
	page, err := paginator.Pages(&paginator.Param{
		DB: db,
		Paging: &paginator.Paging{
			Page:    filter.Page,
			OrderBy: []string{filter.OrderBy},
			Limit:   filter.Limit,
			ShowSQL: false,
		},
	}, &calendarManagements)
	if err != nil {
		return nil, nil, err
	}

	return calendarManagements, page, nil
}

func (r repository) Create(ctx *gin.Context, calendarManagement models.CalendarManagements) (models.CalendarManagements, error) {
	db := r.DB
	err := db.Create(&calendarManagement).Error
	if err != nil {
		return models.CalendarManagements{}, err
	}

	return calendarManagement, nil
}

func (r repository) Update(ctx *gin.Context, calendarManagement models.CalendarManagements) (models.CalendarManagements, error) {
	db := r.DB
	err := db.Save(&calendarManagement).Error
	return calendarManagement, err
}

func (r repository) Delete(ctx *gin.Context, id string) error {
	calendarManagement, notFound, err := r.GetDetail(ctx, id)
	if notFound {
		return fmt.Errorf("calendar management not found")
	} else if err != nil {
		return err
	}

	db := r.DB
	err = db.Delete(&calendarManagement, id).Error
	return err
}
