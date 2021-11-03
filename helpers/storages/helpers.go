package storages

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"time"
)

//generate path
func (base *StorageBase) generatePath(sType string) (string, string, error) {
	ctime := time.Now().Local()
	datePath := ctime.Format("2006/01/02/")

	var storagePath string
	if base.s3Enabled {
		storagePath = os.Getenv("S3_STORAGE_PATH")
	} else if base.gcsEnabled {
		storagePath = os.Getenv("GCS_LOCATION")
	} else {
		storagePath = os.Getenv("LOCAL_STORAGE_PATH")
	}

	filePath := fmt.Sprintf("%s/%s/%s", storagePath, sType, datePath)

	if !base.s3Enabled && !base.gcsEnabled {
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
