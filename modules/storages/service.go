package storages

import (
	"fmt"
	"mime/multipart"
	"os"
	"strings"

	"github.com/fari-99/go-helper/storages"
	"github.com/gin-gonic/gin"

	"go-api/constant"
	"go-api/helpers"
	"go-api/modules/models"
)

type Service interface {
	GetDetail(ctx *gin.Context, storageID uint64) (storageModel *models.Storages, notFound bool, err error)
	Uploads(ctx *gin.Context, form *multipart.Form) ([]models.Storages, error)
	CreateStorage(ctx *gin.Context, storageModel models.Storages) (*models.Storages, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return service{repo: repo}
}

func (s service) GetDetail(ctx *gin.Context, storageID uint64) (*models.Storages, bool, error) {
	return s.repo.GetDetail(ctx, storageID)
}

func (s service) CreateStorage(ctx *gin.Context, storageModel models.Storages) (*models.Storages, error) {
	results, err := s.repo.Create(ctx, []models.Storages{storageModel})
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("failed to save storage record")
	}

	return &results[0], nil
}

func (s service) Uploads(ctx *gin.Context, form *multipart.Form) ([]models.Storages, error) {
	uuid, _ := ctx.Get("uuid")
	currentUser, _ := helpers.GetCurrentUser(ctx, uuid.(string))

	formFile := form.File
	fileType := form.Value["file_types"]

	var storageModels []models.Storages
	for _, files := range formFile {
		for _, file := range files {
			storageBase := newStorageBase(file, fileType[0])
			storageData, err := storageBase.UploadFiles()
			if err != nil {
				return nil, err
			}

			storageModel := models.Storages{
				Type:             storageData.Type,
				Path:             storageData.Path,
				Filename:         storageData.Filename,
				Mime:             storageData.Mime,
				OriginalFilename: storageData.OriginalFilename,
				Status:           constant.StatusActive,
				CreatedBy:        currentUser.ID,
			}

			storageModels = append(storageModels, storageModel)
		}
	}

	if len(storageModels) == 0 {
		return nil, fmt.Errorf("failed to upload your files, please try again")
	}

	results, err := s.repo.Create(ctx, storageModels)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// newStorageBase selects the storage backend from STORAGE_DRIVER (local|s3|gcs).
// Unset or unknown values fall back to local disk.
func newStorageBase(file *multipart.FileHeader, fileType string) *storages.StorageBase {
	storageBase := storages.NewStorageBase(file, fileType)

	switch strings.ToLower(os.Getenv("STORAGE_DRIVER")) {
	case "s3":
		storageBase.SetAwsS3(nil)
	case "gcs":
		storageBase.SetGoogleGCS(nil)
	}

	return storageBase
}
