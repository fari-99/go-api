package storages

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"image/jpeg"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
)

func (base *StorageBase) s3Session() (sessionConfig *session.Session, err error) {
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

func (base *StorageBase) s3Upload(contentTypeData FileData, scaled int, file multipart.File) error {
	sessionCfg, err := base.s3Session()
	if err != nil {
		return fmt.Errorf("failed create session, err := %s", err.Error())
	}

	uploader := s3manager.NewUploader(sessionCfg)

	// create temp file
	fileTemp, err := ioutil.TempFile(os.TempDir(), "prefix")
	if err != nil {
		return fmt.Errorf("bad AWS credentials, err := %s", err.Error())
	}

	if contentTypeData.IsImage {
		// encode all image.Image to jpeg
		// change all image mime to image/jpegjpeg
		var opt jpeg.Options
		opt.Quality = scaled
		err = jpeg.Encode(fileTemp, contentTypeData.ImageFile, &opt)
		contentTypeData.ContentType = "image/jpeg"

		if err != nil {
			return fmt.Errorf("encode image failed -s3-, err := %s", err.Error())
		}
	} else {
		_, err = io.Copy(fileTemp, file)
		if err != nil {
			return fmt.Errorf("error copying data -s3-, err := %s", err.Error())
		}
	}

	_, err = fileTemp.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("error seek file aws to start, err := %s", err.Error())
	}

	params := &s3manager.UploadInput{
		Bucket:      aws.String(os.Getenv("S3_BUCKET")),
		Key:         aws.String(contentTypeData.StoragePath + contentTypeData.Filename),
		Body:        fileTemp,
		ContentType: aws.String(contentTypeData.ContentType),
	}

	_, err = uploader.Upload(params)

	if err != nil {
		return fmt.Errorf("upload to S3 failed, err := %s", err.Error())
	}

	return nil
}

func (base *StorageBase) s3GetFile(storageType, storagePath, filename string) (files *os.File, err error) {
	// Storage on S3
	sessionCfg, err := base.s3Session()
	if err != nil {
		return nil, fmt.Errorf("failed create session, err := %s", err.Error())
	}

	downloader := s3manager.NewDownloader(sessionCfg)
	fileTemp, err := ioutil.TempFile(os.TempDir(), "prefix")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file, err := %s", err.Error())
	}

	filePath := os.Getenv("STORAGE_PATH") + "/" + storageType + storagePath + filename
	_, err = downloader.Download(fileTemp, &s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("S3_BUCKET")),
		Key:    aws.String(filePath),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download file, %v", err)
	}

	return fileTemp, nil
}
