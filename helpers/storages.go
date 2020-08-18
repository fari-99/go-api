package helpers

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/aws/aws-sdk-go/service/s3"
	"go-api/constant"
	"go-api/models"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/kataras/iris"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type FileData struct {
	IsImage     bool
	Extension   string
	ImageFile   image.Image
	ContentType string
}

type StorageBase struct {
	fileInput *multipart.FileHeader
	fileType  string

	ctx       iris.Context
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

func (base *StorageBase) SetCtx(ctx iris.Context) *StorageBase {
	base.ctx = ctx
	return base
}

func (base *StorageBase) S3Session() (sessionConfig *session.Session, err error) {
	awsAccessKey := aws.String(os.Getenv("S3_ACCESS_KEY"))
	awsSecretKey := aws.String(os.Getenv("S3_SECRET_KEY"))
	token := ""

	credential := credentials.NewStaticCredentials(*awsAccessKey, *awsSecretKey, token)
	_, err = credential.Get()
	if err != nil {
		err = fmt.Errorf("bad AWS credentials, err := %s", err.Error())
		return
	}

	cfg := aws.NewConfig().
		WithRegion(os.Getenv("S3_REGION")).
		WithCredentials(credential)

	sessionCfg, err := session.NewSession(cfg)
	if err != nil {
		err = fmt.Errorf("failed create session, err := %s", err.Error())
		return nil, err
	}

	return sessionCfg, nil
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

	if base.s3Enabled {
		sessionCfg, err := base.S3Session()
		if err != nil {
			err = fmt.Errorf("failed create session, err := %s", err.Error())
			return err
		}

		uploader := s3manager.NewUploader(sessionCfg)

		// create temp file
		fileAws, err := ioutil.TempFile(os.TempDir(), "prefix")
		if err != nil {
			err = fmt.Errorf("bad AWS credentials, err := %s", err.Error())
			return err
		}

		if contentTypeData.IsImage {
			// encode all image.Image to jpeg
			// change all image mime to image/jpegjpeg
			var opt jpeg.Options
			opt.Quality = scaled
			err = jpeg.Encode(fileAws, contentTypeData.ImageFile, &opt)
			contentTypeData.ContentType = "image/jpeg"

			if err != nil {
				err = fmt.Errorf("encode image failed, err := %s", err.Error())
				return err
			}
		} else {
			_, err = io.Copy(fileAws, file)
			if err != nil {
				err = fmt.Errorf("error copying data, err := %s", err.Error())
				return err
			}
		}

		_, err = fileAws.Seek(0, 0)
		if err != nil {
			err = fmt.Errorf("bad AWS credentials, err := %s", err.Error())
			return err
		}

		params := &s3manager.UploadInput{
			Bucket:      aws.String(os.Getenv("S3_BUCKET")),
			Key:         aws.String(storagePath + fileName),
			Body:        fileAws,
			ContentType: aws.String(contentTypeData.ContentType),
		}
		_, err = uploader.Upload(params)

		if err != nil {
			err = fmt.Errorf("upload to S3 failed, err := %s", err.Error())
			return err
		}
	} else {
		log.Printf(storagePath + fileName)
		// setup new file
		out, err := os.OpenFile(storagePath+fileName, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			err = fmt.Errorf("temporary file not created, err := %s", err.Error())
			return err
		}

		defer out.Close()

		if contentTypeData.IsImage {
			// encode all image.Image to jpeg
			// change all image mime to image/jpegjpeg
			var opt jpeg.Options
			opt.Quality = scaled
			err = jpeg.Encode(out, contentTypeData.ImageFile, &opt)
			contentTypeData.ContentType = "image/jpeg"

			if err != nil {
				err = fmt.Errorf("encode image failed, err := %s", err.Error())
				return err
			}
		} else {
			_, err = io.Copy(out, file)
			if err != nil {
				return err
			}
		}
	}

	storageModel := models.Storages{
		Type:             fileType,
		Path:             datePath,
		Filename:         fileName,
		Mime:             contentTypeData.ContentType,
		OriginalFilename: fileHeader.Filename,
		CreatedBy:        0,
		Status:           constant.StatusActive,
	}

	if err = base.db.Create(&storageModel).Error; err != nil {
		err = fmt.Errorf("failed save storage data, err := %s", err.Error())
		return
	}

	return
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

func (base *StorageBase) GetFiles(storageModel models.Storages) (files *os.File, err error) {
	var storagePath string
	if !base.s3Enabled {
		storagePath = os.Getenv("LOCAL_STORAGE_PATH") + "/" + storageModel.Type + storageModel.Path + storageModel.Filename
		file, err := os.Open(storagePath)
		if err != nil {
			return nil, fmt.Errorf("error open file, %s", err.Error())
		}

		return file, nil
	}

	// Storage on S3
	sessionCfg, err := base.S3Session()
	if err != nil {
		err = fmt.Errorf("failed create session, err := %s", err.Error())
		return nil, err
	}

	downloader := s3manager.NewDownloader(sessionCfg)
	fileAws, err := ioutil.TempFile(os.TempDir(), "prefix")
	if err != nil {
		err = fmt.Errorf("bad AWS credentials, err := %s", err.Error())
		return nil, err
	}

	storagePath = os.Getenv("STORAGE_PATH") + "/" + storageModel.Type + storageModel.Path + storageModel.Filename
	_, err = downloader.Download(fileAws, &s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("S3_BUCKET")),
		Key:    aws.String(storagePath),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download file, %v", err)
	}

	return fileAws, nil
}

//generate path
func (base *StorageBase) generatePath(sType string) (string, string, error) {
	ctime := time.Now().Local()
	datePath := ctime.Format("2006/01/02/")

	var storagePath string
	if base.s3Enabled {
		storagePath = os.Getenv("STORAGE_PATH")
	} else {
		storagePath = os.Getenv("LOCAL_STORAGE_PATH")
	}

	filePath := fmt.Sprintf("%s/%s/%s", storagePath, sType, datePath)

	if !base.s3Enabled {
		err := os.MkdirAll(filePath, 0711)
		if err != nil {
			return filePath, datePath, err
		}
	}

	return filePath, datePath, nil
}

// generate filename
func (base *StorageBase) generateName(tmpName string, ext string) string {
	ctime := time.Now().Local()
	filename := tmpName + ctime.Format(time.UnixDate)
	hash := md5.Sum([]byte(filename))

	// encode hash to string
	newName := hex.EncodeToString(hash[:])

	return newName + ext
}

func (base *StorageBase) contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
