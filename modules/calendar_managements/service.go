package calendar_managements

import (
	"fmt"

	paginator "github.com/dmitryburov/gorm-paginator"
	"github.com/gin-gonic/gin"

	"go-api/constant"
	"go-api/helpers"
	"go-api/modules/models"
)

type Service interface {
	GetDetail(ctx *gin.Context, id string) (models.CalendarManagements, bool, error)
	GetList(ctx *gin.Context, filter RequestListFilter) ([]models.CalendarManagements, *paginator.Pagination, error)
	Create(ctx *gin.Context, input CreateCalendarManagementRequest) (models.CalendarManagements, error)
	Update(ctx *gin.Context, id string, input UpdateCalendarManagementRequest) (models.CalendarManagements, error)
	UpdateStatus(ctx *gin.Context, id string, status int8) error
	Delete(ctx *gin.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return service{repo: repo}
}

func (s service) GetDetail(ctx *gin.Context, id string) (models.CalendarManagements, bool, error) {
	return s.repo.GetDetail(ctx, id)
}

func (s service) GetList(ctx *gin.Context, filter RequestListFilter) ([]models.CalendarManagements, *paginator.Pagination, error) {
	return s.repo.GetList(ctx, filter)
}

func (s service) Create(ctx *gin.Context, input CreateCalendarManagementRequest) (models.CalendarManagements, error) {
	uuidSession, _ := ctx.Get("uuid")
	currentUser, _ := helpers.GetCurrentUser(ctx, uuidSession.(string))

	model := models.CalendarManagements{
		CountryID:    models.IDType(input.CountryID),
		Dates:        input.Date,
		Descriptions: input.Descriptions,
		CreatedBy:    currentUser.ID,
		Status:       constant.StatusActive,
	}

	return s.repo.Create(ctx, model)
}

func (s service) Update(ctx *gin.Context, id string, input UpdateCalendarManagementRequest) (models.CalendarManagements, error) {
	model, isExists, err := s.GetDetail(ctx, id)
	if !isExists {
		return models.CalendarManagements{}, fmt.Errorf("calendar management not found")
	} else if err != nil {
		return models.CalendarManagements{}, err
	}

	model.CountryID = models.IDType(input.CountryID)
	model.Dates = input.Date
	model.Descriptions = input.Descriptions

	return s.repo.Update(ctx, model)
}

func (s service) UpdateStatus(ctx *gin.Context, id string, status int8) error {
	model, isExists, err := s.GetDetail(ctx, id)
	if !isExists {
		return fmt.Errorf("calendar management not found")
	} else if err != nil {
		return err
	}

	model.Status = status

	_, err = s.repo.Update(ctx, model)
	return err
}

func (s service) Delete(ctx *gin.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
