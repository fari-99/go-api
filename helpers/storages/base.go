package storages

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go-api/constant"
	"go-api/modules/models"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
)

type FileData struct {
	IsImage     bool
	Extension   string
	ImageFile   image.Image
	ContentType string
	StoragePath string
	Filename    string
}

type StorageBase struct {
	fileInput *multipart.FileHeader
	fileType  string

	ctx       *gin.Context
	db        *gorm.DB
	s3Enabled bool
}

func NewStorageBase(fileHeader *multipart.FileHeader, fileType string, db *gorm.DB) *StorageBase {
	s3Enable, _ := strconv.ParseBool(os.Getenv("S3_ENABLE"))

	storageBase := &StorageBase{
		fileInput: fileHeader,
		fileType:  fileType,
		s3Enabled: s3Enable,
		db:        db,
	}

	return storageBase
}

func (base *StorageBase) SetCtx(ctx *gin.Context) *StorageBase {
	base.ctx = ctx
	return base
}

func (base *StorageBase) UploadFiles() (err error) {
	fileHeader := base.fileInput
	fileType := base.fileType

	file, err := fileHeader.Open()
	if err != nil {
		return
	}

	defer file.Close()

	var scaled = 80
	val := os.Getenv("NON_SCALED_TYPE")
	vals := strings.Split(val, ",")

	if base.contains(vals, fileType) == true {
		scaled = 100
	}

	contentTypeData, err := base.getFileData(fileHeader)
	if err != nil {
		return
	}

	storagePath, datePath, err := base.generatePath(fileType)
	if err != nil {
		return
	}

	// Generate hash
	fileName := base.generateName(fileHeader.Filename, contentTypeData.Extension)

	contentTypeData.StoragePath = storagePath
	contentTypeData.Filename = fileName

	if base.s3Enabled {
		err = base.S3Upload(contentTypeData, scaled, file)
	} else {
		err = base.LocalUpload(contentTypeData, scaled, file)
	}

	if err != nil {
		return err
	}

	storageModel := models.Storages{
		Type:             fileType,
		Path:             datePath,
		Filename:         fileName,
		Mime:             contentTypeData.ContentType,
		OriginalFilename: fileHeader.Filename,
		CreatedBy:        1,
		Status:           constant.StatusActive,
	}

	if err = base.db.Create(&storageModel).Error; err != nil {
		err = fmt.Errorf("failed save storage data, err := %s", err.Error())
		return
	}

	return
}

func (base *StorageBase) GetFiles(storageModel models.Storages) (files *os.File, err error) {
	if base.s3Enabled {
		return base.S3GetFile(storageModel)
	} else {
		return base.LocalGetFile(storageModel)
	}
}

func (base *StorageBase) getFileData(fileHeader *multipart.FileHeader) (contentTypeData FileData, err error) {
	file, err := fileHeader.Open()
	if err != nil {
		return
	}

	defer file.Close()

	buffer := make([]byte, 1024)
	_, err = file.Read(buffer)
	if err != nil {
		err = fmt.Errorf("file could not be read, err := %s", err.Error())
		return
	}

	_, _ = file.Seek(0, 0)
	contentType := http.DetectContentType(buffer)

	var img image.Image
	var isImage = true
	var ext string

	switch contentType {
	case "image/png":
		img, err = png.Decode(file)
		ext = ".jpg"
	case "image/gif":
		img, err = gif.Decode(file)
		ext = ".jpg"
	case "image/jpeg":
		img, err = jpeg.Decode(file)
		ext = ".jpg"
	case "image/jpg":
		img, err = jpeg.Decode(file)
		ext = ".jpg"
	default:
		isImage = false
		// Get file extension
		ext = path.Ext(fileHeader.Filename)
	}

	contentTypeData = FileData{
		IsImage:     isImage,
		Extension:   ext,
		ImageFile:   img,
		ContentType: contentType,
	}

	return
}
