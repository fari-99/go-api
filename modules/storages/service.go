package storages

import (
	"github.com/gin-gonic/gin"
	"go-api/helpers"
	"go-api/helpers/storages"
	"go-api/modules/models"
	"mime/multipart"
)

type Service interface {
	GetDetail(ctx *gin.Context, storageID int64) (storageModel *models.Storages, notFound bool, err error)
	Uploads(ctx *gin.Context, form *multipart.Form) ([]models.Storages, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return service{repo: repo}
}

func (s service) GetDetail(ctx *gin.Context, storageID int64) (*models.Storages, bool, error) {
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

			if currentUser != nil {
				storageBase.SetCreatedBy(currentUser.ID)
			}

			storageModel, err := storageBase.UploadFiles()
			if err != nil {
				return nil, err
			}

			savedModel, err := s.repo.Create(ctx, *storageModel)
			if err != nil {
				continue
			}

			storageModels = append(storageModels, savedModel)
		}
	}

	return storageModels, nil
}
