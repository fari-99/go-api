package file_managements

import (
	"context"
	"errors"

	paginator "github.com/dmitryburov/gorm-paginator"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"go-api/constant"
	"go-api/helpers/activitylog"
	"go-api/modules/configs"
	"go-api/modules/models"
)

// ErrTransitionLost means the row was no longer in the expected status (cancelled, taken by
// another worker, deleted, ...), so nothing was written.
var ErrTransitionLost = errors.New("file management status changed")

type Repository interface {
	GetStorage(ctx context.Context, storageID uint64) (*models.Storages, bool, error)
	GetDetail(ctx context.Context, fileManagementID uint64) (*models.FileManagements, bool, error)
	GetList(ctx context.Context, filter ListFilter) ([]models.FileManagements, *paginator.Pagination, error)
	GetMetas(ctx context.Context, fileManagementID uint64) ([]models.FileManagementMeta, error)
	// GetActive returns the user's new / on progress jobs for a storage and action
	GetActive(ctx context.Context, createdBy, storageID uint64, action string) ([]models.FileManagements, error)
	CountActive(ctx context.Context, createdBy uint64) (int64, error)
	// Create inserts the job and its metas in one transaction
	Create(ctx context.Context, fileManagement *models.FileManagements, metas []models.FileManagementMeta, actorID uint64) error
	// Transition updates the job only if it is still in status from, and upserts metas in the same
	// transaction. Returns ErrTransitionLost when the status already changed.
	Transition(ctx context.Context, fileManagementID uint64, from int8, updates map[string]interface{}, metas []models.FileManagementMeta, actorID uint64) error
	// Delete soft deletes a job in a terminal status with its metas.
	Delete(ctx context.Context, fileManagementID uint64, actorID uint64) error
}

type repository struct {
	*configs.DI
}

func NewRepository(di *configs.DI) Repository {
	return repository{DI: di}
}

var activeStatuses = []int8{constant.FileManagementNew, constant.FileManagementOnProgress}
var terminalStatuses = []int8{constant.FileManagementCompleted, constant.FileManagementFailed, constant.FileManagementCancelled}

func (r repository) db(ctx context.Context, actorID uint64) *gorm.DB {
	return activitylog.WithActor(r.DB.WithContext(ctx), actorID)
}

func (r repository) GetStorage(ctx context.Context, storageID uint64) (*models.Storages, bool, error) {
	var storageModel models.Storages
	err := r.DB.WithContext(ctx).Where("id = ?", storageID).First(&storageModel).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, true, nil
	} else if err != nil {
		return nil, false, err
	}

	return &storageModel, false, nil
}

func (r repository) GetDetail(ctx context.Context, fileManagementID uint64) (*models.FileManagements, bool, error) {
	var fileManagement models.FileManagements
	err := r.DB.WithContext(ctx).Where("id = ?", fileManagementID).First(&fileManagement).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, true, nil
	} else if err != nil {
		return nil, false, err
	}

	return &fileManagement, false, nil
}

// GetList returns the user's jobs, newest first.
func (r repository) GetList(ctx context.Context, filter ListFilter) ([]models.FileManagements, *paginator.Pagination, error) {
	db := r.DB.WithContext(ctx).Where("created_by = ?", filter.CreatedBy)
	if filter.Status != 0 {
		db = db.Where("status = ?", filter.Status)
	}
	if filter.Action != "" {
		db = db.Where("action = ?", filter.Action)
	}
	if filter.StorageID != 0 {
		db = db.Where("storage_id = ?", filter.StorageID)
	}

	var fileManagements []models.FileManagements
	result, err := paginator.Pages(&paginator.Param{
		DB: db,
		Paging: &paginator.Paging{
			Page:    filter.Page,
			Limit:   filter.Limit,
			OrderBy: []string{"id desc"},
		},
	}, &fileManagements)
	if err != nil {
		return nil, nil, err
	}

	return fileManagements, result, nil
}

func (r repository) GetMetas(ctx context.Context, fileManagementID uint64) ([]models.FileManagementMeta, error) {
	var metas []models.FileManagementMeta
	err := r.DB.WithContext(ctx).
		Where("file_management_id = ?", fileManagementID).
		Order("id asc").
		Find(&metas).Error
	return metas, err
}

func (r repository) GetActive(ctx context.Context, createdBy, storageID uint64, action string) ([]models.FileManagements, error) {
	var fileManagements []models.FileManagements
	err := r.DB.WithContext(ctx).
		Where("created_by = ? AND storage_id = ? AND action = ? AND status IN ?", createdBy, storageID, action, activeStatuses).
		Order("id asc").
		Find(&fileManagements).Error
	return fileManagements, err
}

func (r repository) CountActive(ctx context.Context, createdBy uint64) (int64, error) {
	var total int64
	err := r.DB.WithContext(ctx).Model(&models.FileManagements{}).
		Where("created_by = ? AND status IN ?", createdBy, activeStatuses).
		Count(&total).Error
	return total, err
}

func (r repository) Create(ctx context.Context, fileManagement *models.FileManagements, metas []models.FileManagementMeta, actorID uint64) error {
	return r.db(ctx, actorID).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(fileManagement).Error; err != nil {
			return err
		}

		for i := range metas {
			metas[i].FileManagementID = fileManagement.ID
		}

		return upsertMetas(tx, metas)
	})
}

func (r repository) Transition(ctx context.Context, fileManagementID uint64, from int8, updates map[string]interface{}, metas []models.FileManagementMeta, actorID uint64) error {
	return r.db(ctx, actorID).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&models.FileManagements{Base: models.Base{ID: models.IDType(fileManagementID)}}).
			Where("status = ?", from).
			Updates(updates)
		if result.Error != nil {
			return result.Error
		} else if result.RowsAffected == 0 {
			return ErrTransitionLost
		}

		return upsertMetas(tx, metas)
	})
}

func (r repository) Delete(ctx context.Context, fileManagementID uint64, actorID uint64) error {
	return r.db(ctx, actorID).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("status IN ?", terminalStatuses).
			Delete(&models.FileManagements{Base: models.Base{ID: models.IDType(fileManagementID)}})
		if result.Error != nil {
			return result.Error
		} else if result.RowsAffected == 0 {
			return ErrTransitionLost
		}

		return tx.Where("file_management_id = ?", fileManagementID).Delete(&models.FileManagementMeta{}).Error
	})
}

// upsertMetas inserts metas, overwriting the value of an existing (file_management_id, group, type, key).
func upsertMetas(tx *gorm.DB, metas []models.FileManagementMeta) error {
	if len(metas) == 0 {
		return nil
	}

	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "file_management_id"}, {Name: "group"}, {Name: "type"}, {Name: "key"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at", "deleted_at"}),
	}).Create(&metas).Error
}
