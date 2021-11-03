package storages

import (
	"fmt"
	"github.com/spf13/cast"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"net/http"
	"os"
	"path"
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

	s3Enabled  bool
	gcsEnabled bool
}

type StorageData struct {
	Type             string `json:"type"`
	Path             string `json:"path"`
	Filename         string `json:"filename"`
	Mime             string `json:"mime"`
	OriginalFilename string `json:"original_filename"`
}

func NewStorageBase(fileHeader *multipart.FileHeader, fileType string) *StorageBase {
	s3Enable := cast.ToBool(os.Getenv("S3_ENABLE"))
	gcsEnable := cast.ToBool(os.Getenv("GCS_ENABLED"))

	storageBase := &StorageBase{
		fileInput:  fileHeader,
		fileType:   fileType,
		s3Enabled:  s3Enable,
		gcsEnabled: gcsEnable,
	}

	return storageBase
}

func (base *StorageBase) UploadFiles() (storageModel *StorageData, err error) {
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
		err = base.s3Upload(contentTypeData, scaled, file)
	} else if base.gcsEnabled {
		err = base.gcsUpload(contentTypeData, scaled, file)
	} else {
		err = base.localUpload(contentTypeData, scaled, file)
	}

	if err != nil {
		return nil, err
	}

	storageModel = &StorageData{
		Type:             fileType,
		Path:             datePath,
		Filename:         fileName,
		Mime:             contentTypeData.ContentType,
		OriginalFilename: fileHeader.Filename,
	}

	return storageModel, nil
}

func (base *StorageBase) GetFiles(storageType, storagePath, filename string) (files *os.File, err error) {
	if base.s3Enabled {
		return base.s3GetFile(storageType, storagePath, filename)
	} else if base.gcsEnabled {
		return base.gcsGetFile(storageType, storagePath, filename)
	} else {
		return base.localGetFile(storageType, storagePath, filename)
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
