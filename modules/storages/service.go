package storages

import (
	"fmt"
	"mime/multipart"

	"github.com/gin-gonic/gin"

	"go-api/constant"
	"go-api/helpers"
	"go-api/helpers/storages"
	"go-api/modules/models"
)

type Service interface {
	GetDetail(ctx *gin.Context, storageID uint64) (storageModel *models.Storages, notFound bool, err error)
	Uploads(ctx *gin.Context, form *multipart.Form) ([]models.Storages, error)
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

func (s service) Uploads(ctx *gin.Context, form *multipart.Form) ([]models.Storages, error) {
	uuid, _ := ctx.Get("uuid")
	currentUser, _ := helpers.GetCurrentUser(uuid.(string))

	formFile := form.File
	fileType := form.Value["file_types"]

	var storageModels []models.Storages
	for _, files := range formFile {
		for _, file := range files {
			storageBase := storages.NewStorageBase(file, fileType[0])
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
