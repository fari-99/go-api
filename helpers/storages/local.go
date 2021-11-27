package storages

import (
	"fmt"
	"image/jpeg"
	"io"
	"log"
	"mime/multipart"
	"os"
)

func (base *StorageBase) localUpload(contentTypeData FileData, scaled int, file multipart.File) error {
	storagePath := contentTypeData.StoragePath
	fileName := contentTypeData.Filename

	log.Printf(storagePath + fileName)
	// setup new file
	out, err := os.OpenFile(storagePath+fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("temporary file not created, err := %s", err.Error())
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
			return fmt.Errorf("encode image failed, err := %s", err.Error())
		}
	} else {
		_, err = io.Copy(out, file)
		if err != nil {
			return fmt.Errorf("error upload file to local, err := %s", err.Error())
		}
	}

	return err
}

func (base *StorageBase) localGetFile(storageType, storagePath, filename string) (files *os.File, err error) {
	filePath := os.Getenv("LOCAL_STORAGE_PATH") + "/" + storageType + storagePath + filename
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error open file, %s", err.Error())
	}

	return file, nil
}
