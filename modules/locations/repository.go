package locations

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"go-api/modules/configs"
	"go-api/modules/models"

	"github.com/gin-gonic/gin"
)

type Repository interface {
	// - Locations
	GetDetailLocation(ctx *gin.Context, locationID string) (locationModel *models.Locations, notFound bool, err error)
	CountLocation(ctx *gin.Context, filter FilterQueryLocations) (int64, error)
	GetAllLocation(ctx *gin.Context, filter FilterQueryLocations, limit, offset int) (locationModels []models.Locations, err error)
	CreateLocation(ctx *gin.Context, input models.Locations) (locationModel *models.Locations, err error)
	UpdateLocation(ctx *gin.Context, locationID string, input models.Locations) (locationModel *models.Locations, err error)
	UpdateStatusLocation(ctx *gin.Context, locationID string, status int8) (locationModel *models.Locations, err error)
	DeleteLocation(ctx *gin.Context, locationID string) error

	// - Location Levels
	GetDetailLocationLevel(ctx *gin.Context, locationID string) (locationModel *models.LocationLevels, notFound bool, err error)
	CountLocationLevel(ctx *gin.Context, filter FilterQueryLocationLevel) (int64, error)
	GetAllLocationLevel(ctx *gin.Context, filter FilterQueryLocationLevel, limit, offset int) (locationModels []models.LocationLevels, err error)
	CreateLocationLevel(ctx *gin.Context, input models.LocationLevels) (locationModel *models.LocationLevels, err error)
	UpdateLocationLevel(ctx *gin.Context, locationLevelID string, input models.LocationLevels) (locationModel *models.LocationLevels, err error)
	UpdateStatusLocationLevel(ctx *gin.Context, locationLevelID string, status int8) (locationModel *models.LocationLevels, err error)
	DeleteLocationLevel(ctx *gin.Context, locationID string) error
}

type repository struct {
	*configs.DI
}

func NewRepository(di *configs.DI) Repository {
	return repository{DI: di}
}

func (r repository) GetDetailLocation(ctx *gin.Context, locationID string) (*models.Locations, bool, error) {
	db := r.DB.WithContext(ctx)

	var locationModel models.Locations
	err := db.Where(&models.Locations{Base: models.Base{ID: models.IDType(locationID)}}).First(&locationModel).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, true, nil
	} else if err != nil {
		return nil, true, err
	}

	return &locationModel, false, nil
}

func (r repository) CountLocation(ctx *gin.Context, filter FilterQueryLocations) (int64, error) {
	db := r.DB.WithContext(ctx).Model(&models.Locations{})

	var count int64
	r.FilterLocations(db, filter, 0, 0)
	err := db.Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r repository) GetAllLocation(ctx *gin.Context, filter FilterQueryLocations, limit, offset int) ([]models.Locations, error) {
	db := r.DB.WithContext(ctx)

	var locationModels []models.Locations
	r.FilterLocations(db, filter, limit, offset)
	err := db.Find(&locationModels).Error
	if err != nil {
		return nil, err
	}

	return locationModels, nil
}

func (r repository) CreateLocation(ctx *gin.Context, locationModel models.Locations) (*models.Locations, error) {
	db := r.DB.WithContext(ctx)

	err := db.Create(&locationModel).Error
	if err != nil {
		return nil, err
	}

	return &locationModel, nil
}

func (r repository) UpdateLocation(ctx *gin.Context, locationID string, input models.Locations) (*models.Locations, error) {
	db := r.DB.WithContext(ctx)

	locationModel, notFound, err := r.GetDetailLocation(ctx, locationID)
	if err != nil {
		return nil, err
	} else if notFound {
		return nil, fmt.Errorf("location not found")
	}

	err = db.Model(locationModel).Updates(input).Error
	if err != nil {
		return nil, err
	}

	return locationModel, nil
}

func (r repository) UpdateStatusLocation(ctx *gin.Context, locationID string, status int8) (*models.Locations, error) {
	db := r.DB.WithContext(ctx)

	locationModel, notFound, err := r.GetDetailLocation(ctx, locationID)
	if err != nil {
		return nil, err
	} else if notFound {
		return nil, fmt.Errorf("location not found")
	}

	err = db.Model(locationModel).Update("status", status).Error
	if err != nil {
		return nil, err
	}

	return locationModel, nil
}

func (r repository) DeleteLocation(ctx *gin.Context, locationID string) error {
	db := r.DB.WithContext(ctx)

	locationModel, _, _ := r.GetDetailLocation(ctx, locationID)
	err := db.Delete(locationModel).Error
	if err != nil {
		return err
	}

	return nil
}

func (r repository) GetDetailLocationLevel(ctx *gin.Context, locationID string) (*models.LocationLevels, bool, error) {
	db := r.DB.WithContext(ctx)

	var locationModel models.LocationLevels
	err := db.Where(&models.LocationLevels{Base: models.Base{ID: models.IDType(locationID)}}).First(&locationModel).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, true, nil
	} else if err != nil {
		return nil, true, err
	}

	return &locationModel, false, nil
}

func (r repository) CountLocationLevel(ctx *gin.Context, filter FilterQueryLocationLevel) (int64, error) {
	db := r.DB.WithContext(ctx).Model(&models.LocationLevels{})

	var count int64
	dbFilterCount := r.FilterLocationLevels(db, filter, 0, 0)
	err := dbFilterCount.Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r repository) GetAllLocationLevel(ctx *gin.Context, filter FilterQueryLocationLevel, limit, offset int) ([]models.LocationLevels, error) {
	db := r.DB.WithContext(ctx)

	var locationModels []models.LocationLevels
	dbFilter := r.FilterLocationLevels(db, filter, limit, offset)
	err := dbFilter.Find(&locationModels).Error
	if err != nil {
		return nil, err
	}

	return locationModels, nil
}

func (r repository) CreateLocationLevel(ctx *gin.Context, locationModel models.LocationLevels) (*models.LocationLevels, error) {
	db := r.DB.WithContext(ctx)

	err := db.Create(&locationModel).Error
	if err != nil {
		return nil, err
	}

	return &locationModel, nil
}

func (r repository) UpdateLocationLevel(ctx *gin.Context, locationLevelID string, input models.LocationLevels) (*models.LocationLevels, error) {
	db := r.DB.WithContext(ctx)

	locationLevelModel, notFound, err := r.GetDetailLocationLevel(ctx, locationLevelID)
	if err != nil {
		return nil, err
	} else if notFound {
		return nil, fmt.Errorf("location level not found")
	}

	err = db.Model(locationLevelModel).Updates(input).Error
	if err != nil {
		return nil, err
	}

	return locationLevelModel, nil
}

func (r repository) UpdateStatusLocationLevel(ctx *gin.Context, locationLevelID string, status int8) (*models.LocationLevels, error) {
	db := r.DB.WithContext(ctx)

	locationLevelModel, notFound, err := r.GetDetailLocationLevel(ctx, locationLevelID)
	if err != nil {
		return nil, err
	} else if notFound {
		return nil, fmt.Errorf("location level not found")
	}

	err = db.Model(locationLevelModel).Update("status", status).Error
	if err != nil {
		return nil, err
	}

	return locationLevelModel, nil
}

func (r repository) DeleteLocationLevel(ctx *gin.Context, locationLevelID string) error {
	db := r.DB.WithContext(ctx)

	locationLevelModel, notFound, err := r.GetDetailLocationLevel(ctx, locationLevelID)
	if err != nil {
		return err
	} else if notFound {
		return fmt.Errorf("location level not found")
	}

	err = db.Delete(locationLevelModel).Error
	if err != nil {
		return err
	}

	return nil
}
