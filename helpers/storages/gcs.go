package storages

import (
	"context"
	"fmt"
	"github.com/spf13/cast"
	"image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"os"
	"time"

	"cloud.google.com/go/storage"
)

func (base *StorageBase) gcsInit() (*storage.Client, error) {
	log.Printf("upload using GCS (Google Cloud Storage)")
	err := os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", os.Getenv("GCS_CREDENTIAL_PATH"))
	if err != nil {
		return nil, err
	}

	client, err := storage.NewClient(context.Background())
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (base *StorageBase) gcsUpload(contentTypeData FileData, scaled int, file multipart.File) error {
	client, err := base.gcsInit()
	if err != nil {
		return err
	}

	// close client after use
	defer client.Close()

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(cast.ToInt(os.Getenv("GCS_TIMEOUT"))))
	defer cancel()

	// create temp file
	fileTemp, err := ioutil.TempFile(os.TempDir(), "prefix")
	if err != nil {
		return fmt.Errorf("bad GCS credentials, err := %s", err.Error())
	}

	if contentTypeData.IsImage {
		// encode all image.Image to jpeg
		// change all image mime to image/jpegjpeg
		var opt jpeg.Options
		opt.Quality = scaled
		err = jpeg.Encode(fileTemp, contentTypeData.ImageFile, &opt)
		contentTypeData.ContentType = "image/jpeg"

		if err != nil {
			return fmt.Errorf("encode image failed -GCS-, err := %s", err.Error())
		}
	} else {
		_, err = io.Copy(fileTemp, file)
		if err != nil {
			return fmt.Errorf("error copying data -GCS-, err := %s", err.Error())
		}
	}

	_, err = fileTemp.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("error seek file GCS to start, err := %s", err.Error())
	}

	// Upload an object with storage.Writer.
	filePath := contentTypeData.StoragePath + contentTypeData.Filename
	wc := client.Bucket(os.Getenv("GCS_BUCKET_NAME")).Object(filePath).NewWriter(ctx)
	if _, err = io.Copy(wc, fileTemp); err != nil {
		return fmt.Errorf("error copying data -gcs-, err := %s", err.Error())
	}

	if err = wc.Close(); err != nil {
		return fmt.Errorf("error close writer -gcs-: %v", err)
	}

	return nil
}

func (base *StorageBase) gcsGetFile(storageType, storagePath, filename string) (*os.File, error) {
	client, err := base.gcsInit()
	if err != nil {
		return nil, err
	}

	// close client after use
	defer client.Close()

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(cast.ToInt(os.Getenv("GCS_TIMEOUT"))))
	defer cancel()

	fileTemp, err := ioutil.TempFile(os.TempDir(), "prefix")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file, err := %s", err.Error())
	}

	filePath := fmt.Sprintf("%s/%s/%s%s", os.Getenv("GCS_LOCATION"), storageType, storagePath, filename)
	rc, err := client.Bucket(os.Getenv("GCS_BUCKET_NAME")).Object(filePath).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("error create new reader -gcs-, err := %s", err.Error())
	}
	defer rc.Close()

	_, err = io.Copy(fileTemp, rc)
	if err != nil {
		return nil, fmt.Errorf("failed to download file, %v", err)
	}

	_, err = fileTemp.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("error seek file temp -gcs- to start, err := %s", err.Error())
	}

	return fileTemp, nil

}
