package locations

import (
	"fmt"

	"go-api/constant"
	"go-api/helpers"
	"go-api/modules/models"

	"github.com/gin-gonic/gin"
)

type Service interface {
	GetDetailLocation(ctx *gin.Context, locationID string) (*models.Locations, bool, error)
	GetAllLocation(ctx *gin.Context, filter FilterQueryLocations) (*helpers.Pages, error)
	CreateLocation(ctx *gin.Context, input RequestCreateLocations) (*models.Locations, error)
	UpdateLocation(ctx *gin.Context, locationID string, input RequestUpdateLocations) (*models.Locations, error)
	UpdateStatusLocation(ctx *gin.Context, locationID string, status int8) (*models.Locations, error)
	DeleteLocation(ctx *gin.Context, locationID string) error

	GetDetailLocationLevel(ctx *gin.Context, locationLevelID string) (*models.LocationLevels, bool, error)
	GetAllLocationLevel(ctx *gin.Context, filter FilterQueryLocationLevel) (*helpers.Pages, error)
	CreateLocationLevel(ctx *gin.Context, input RequestCreateLocationLevel) (*models.LocationLevels, error)
	UpdateLocationLevel(ctx *gin.Context, locationLevelID string, input RequestUpdateLocationLevel) (*models.LocationLevels, error)
	UpdateStatusLocationLevel(ctx *gin.Context, locationLevelID string, status int8) (*models.LocationLevels, error)
	DeleteLocationLevel(ctx *gin.Context, locationLevelID string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return service{repo: repo}
}

func (s service) getCompleteName(ctx *gin.Context, inputName, parentID string) (string, error) {
	// get parent data
	completeName := inputName
	if parentID != "" {
		parentLocationModel, notFound, err := s.repo.GetDetailLocation(ctx, parentID)
		if err != nil {
			return "", err
		} else if notFound {
			return "", fmt.Errorf("parent data not found")
		}

		completeName = parentLocationModel.CompleteName + ", " + inputName
	}

	return completeName, nil
}

func (s service) GetDetailLocation(ctx *gin.Context, locationID string) (*models.Locations, bool, error) {
	return s.repo.GetDetailLocation(ctx, locationID)
}

func (s service) GetAllLocation(ctx *gin.Context, filter FilterQueryLocations) (*helpers.Pages, error) {
	count, err := s.repo.CountLocation(ctx, filter)
	if err != nil {
		return nil, err
	}

	pages := helpers.NewFromRequest(ctx.Request, int(count))
	items, err := s.repo.GetAllLocation(ctx, filter, pages.Limit(), pages.Offset())
	if err != nil {
		return nil, err
	}

	for idx, item := range items {
		level, notFound, err := s.repo.GetDetailLocationLevel(ctx, string(item.LevelID))
		if err != nil {
			return nil, err
		} else if notFound {
			return nil, fmt.Errorf("location level not found")
		}

		items[idx].Levels = *level
	}

	pages.Items = items
	return pages, nil
}

func (s service) CreateLocation(ctx *gin.Context, input RequestCreateLocations) (*models.Locations, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	// check if level id exists
	_, notFound, err := s.repo.GetDetailLocationLevel(ctx, input.LevelID)
	if err != nil {
		return nil, err
	} else if notFound {
		return nil, fmt.Errorf("location level not found")
	}

	completeName, err := s.getCompleteName(ctx, input.Name, input.ParentID)
	if err != nil {
		return nil, err
	}

	locationModel := models.Locations{
		ParentID:     models.IDType(input.ParentID),
		Code:         input.Code,
		Name:         input.Name,
		CompleteName: completeName,
		LevelID:      models.IDType(input.LevelID),
		Status:       constant.StatusActive,
	}

	savedModel, err := s.repo.CreateLocation(ctx, locationModel)
	return savedModel, err
}

func (s service) UpdateLocation(ctx *gin.Context, locationID string, input RequestUpdateLocations) (*models.Locations, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	completeName, err := s.getCompleteName(ctx, input.Name, input.ParentID)
	if err != nil {
		return nil, err
	}

	locationModel := models.Locations{
		ParentID:     models.IDType(input.ParentID),
		Code:         input.Code,
		Name:         input.Name,
		CompleteName: completeName,
		LevelID:      models.IDType(input.LevelID),
		Status:       input.Status,
	}

	// update
	newLocationModel, err := s.repo.UpdateLocation(ctx, locationID, locationModel)
	if err != nil {
		return nil, err
	}

	return newLocationModel, nil
}

func (s service) UpdateStatusLocation(ctx *gin.Context, locationID string, status int8) (*models.Locations, error) {
	_, notFound, err := s.repo.GetDetailLocation(ctx, locationID)
	if err != nil {
		return nil, err
	} else if notFound {
		return nil, fmt.Errorf("location data not found")
	}

	locationModel, err := s.repo.UpdateStatusLocation(ctx, locationID, status)
	if err != nil {
		return nil, err
	}

	return locationModel, nil
}

func (s service) DeleteLocation(ctx *gin.Context, locationID string) error {
	return s.repo.DeleteLocation(ctx, locationID)
}

func (s service) GetDetailLocationLevel(ctx *gin.Context, locationID string) (*models.LocationLevels, bool, error) {
	return s.repo.GetDetailLocationLevel(ctx, locationID)
}

func (s service) GetAllLocationLevel(ctx *gin.Context, filter FilterQueryLocationLevel) (*helpers.Pages, error) {
	count, err := s.repo.CountLocationLevel(ctx, filter)
	if err != nil {
		return nil, err
	}

	pages := helpers.NewFromRequest(ctx.Request, int(count))
	items, err := s.repo.GetAllLocationLevel(ctx, filter, pages.Limit(), pages.Offset())
	if err != nil {
		return nil, err
	}

	pages.Items = items
	return pages, nil
}

func (s service) CreateLocationLevel(ctx *gin.Context, input RequestCreateLocationLevel) (*models.LocationLevels, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	locationModel := models.LocationLevels{
		NeedParentID: *input.NeedParentID,
		Name:         input.Name,
		Status:       input.Status,
	}

	savedModel, err := s.repo.CreateLocationLevel(ctx, locationModel)
	return savedModel, err
}

func (s service) UpdateLocationLevel(ctx *gin.Context, locationID string, input RequestUpdateLocationLevel) (*models.LocationLevels, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	locationModel := models.LocationLevels{
		NeedParentID: *input.NeedParentID,
		Name:         input.Name,
		Status:       input.Status,
	}

	// update
	newLocationModel, err := s.repo.UpdateLocationLevel(ctx, locationID, locationModel)
	if err != nil {
		return nil, err
	}

	return newLocationModel, nil
}

func (s service) UpdateStatusLocationLevel(ctx *gin.Context, locationID string, status int8) (*models.LocationLevels, error) {
	_, notFound, err := s.repo.GetDetailLocationLevel(ctx, locationID)
	if err != nil {
		return nil, err
	} else if notFound {
		return nil, fmt.Errorf("location data not found")
	}

	locationModel, err := s.repo.UpdateStatusLocationLevel(ctx, locationID, status)
	if err != nil {
		return nil, err
	}

	return locationModel, nil
}

func (s service) DeleteLocationLevel(ctx *gin.Context, locationID string) error {
	return s.repo.DeleteLocationLevel(ctx, locationID)
}
