package security_cameras

import (
	"fmt"

	paginator "github.com/dmitryburov/gorm-paginator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"go-api/helpers"
	"go-api/modules/models"
	"go-api/pkg/go2rtc_helper"
)

type Service interface {
	GetDetail(ctx *gin.Context, id string) (*models.SecurityCameras, bool, error)
	GetList(ctx *gin.Context, filter RequestListFilter) ([]models.SecurityCameras, *paginator.Pagination, error)
	Create(ctx *gin.Context, model models.SecurityCameras) (*models.SecurityCameras, error)
	Update(ctx *gin.Context, id string, model models.SecurityCameras) (*models.SecurityCameras, error)
	Delete(ctx *gin.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return service{repo: repo}
}

func (s service) GetDetail(ctx *gin.Context, id string) (*models.SecurityCameras, bool, error) {
	model, isNotExists, err := s.repo.GetDetail(ctx, id)
	if err != nil {
		return nil, false, err
	} else if isNotExists {
		return nil, false, nil
	}

	stream := s.GetSource(*model)

	helper, err := go2rtc_helper.New()
	if err != nil {
		return nil, false, err
	}

	if isExists, err := helper.StreamExists(stream.Name); err != nil {
		return nil, false, err
	} else if !isExists {
		err = helper.AddStream(stream.Name, stream.Source)
		if err != nil {
			return nil, false, err
		}
	}

	model.Stream = stream
	return model, true, nil
}

func (s service) GetList(ctx *gin.Context, filter RequestListFilter) ([]models.SecurityCameras, *paginator.Pagination, error) {
	listModel, paginatorData, err := s.repo.GetList(ctx, filter)
	if err != nil {
		return nil, nil, err
	}

	helper, err := go2rtc_helper.New()
	if err != nil {
		return nil, nil, err
	}

	for idx, model := range listModel {
		stream := s.GetSource(model)
		if isExists, err := helper.StreamExists(stream.Name); err != nil {
			return nil, nil, err
		} else if !isExists {
			err = helper.AddStream(stream.Name, stream.Source)
			if err != nil {
				return nil, nil, err
			}
		}

		listModel[idx].Stream = stream
	}

	return listModel, paginatorData, nil
}

func (s service) Create(ctx *gin.Context, model models.SecurityCameras) (*models.SecurityCameras, error) {
	uuidSession, _ := ctx.Get("uuid")
	currentUser, _ := helpers.GetCurrentUser(ctx, uuidSession.(string))

	model.Uuid = uuid.New().String()
	model.UserID = string(currentUser.ID)

	stream := s.GetSource(model)

	// add to go2rtc
	helper, err := go2rtc_helper.New()
	if err != nil {
		return nil, err
	}

	err = helper.AddStream(stream.Name, stream.Source)
	if err != nil {
		return nil, err
	}

	// add to database
	return s.repo.Create(ctx, model)
}

func (s service) Update(ctx *gin.Context, id string, input models.SecurityCameras) (*models.SecurityCameras, error) {
	model, isExists, err := s.GetDetail(ctx, id)
	if err != nil {
		return nil, err
	} else if !isExists {
		return nil, fmt.Errorf("security cameras %s does not exist", id)
	} else if model == nil {
		return nil, fmt.Errorf("security cameras %s does not exist", id)
	}

	currentStream := s.GetSource(*model)

	// update to go2rtc
	helper, err := go2rtc_helper.New()
	if err != nil {
		return nil, err
	}

	if ok, err := helper.StreamExists(currentStream.Name); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("security cameras %s does not exist", id)
	}

	model.Url = input.Url
	model.Username = input.Username
	model.Password = input.Password
	model.Descriptions = input.Descriptions

	updateStream := s.GetSource(*model)
	err = helper.UpdateStream(updateStream.Name, updateStream.Source)
	if err != nil {
		return nil, err
	}

	return s.repo.Update(ctx, *model)
}

func (s service) Delete(ctx *gin.Context, id string) error {
	model, isExists, err := s.GetDetail(ctx, id)
	if err != nil {
		return err
	} else if !isExists {
		return fmt.Errorf("security camera id [%s] is not exists", id)
	}

	// delete to go2rtc
	helper, err := go2rtc_helper.New()
	if err != nil {
		return err
	}

	err = helper.DeleteStream(model.Stream.Name)
	if err != nil {
		return err
	}

	// delete to database
	return s.repo.Delete(ctx, id)
}

func (s service) GetSource(model models.SecurityCameras) go2rtc_helper.Stream {
	input := go2rtc_helper.InputGo2RTC{
		Name:     fmt.Sprintf("%s [%s]", model.Name, model.Uuid),
		Url:      model.Url,
		Username: model.Username,
		Password: model.Password,
	}

	sourceUrl := go2rtc_helper.GenerateUrl(input)
	return go2rtc_helper.Stream{
		Name:   input.Name,
		Source: sourceUrl,
	}
}
