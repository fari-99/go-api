package storages

import (
	"bytes"
	"fmt"
	"go-api/helpers"
	"go-api/helpers/crypts"
	"go-api/helpers/storages"
	"go-api/modules/configs"
	"go-api/modules/models"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/nfnt/resize"
	"github.com/spf13/cast"
)

type StorageController struct {
	*configs.DI
}

func (controller *StorageController) UploadAction(ctx *gin.Context) {
	err := ctx.Request.ParseMultipartForm(8 << 20) // 8 MB
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	form := ctx.Request.MultipartForm
	files := form.File["files[]"]
	for _, file := range files {
		err = storages.NewStorageBase(file, "test", controller.DB).SetCtx(ctx).UploadFiles()
		if err != nil {
			helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
			return
		}
	}

	helpers.NewResponse(ctx, http.StatusOK, "Success upload files")
	return
}

func (controller *StorageController) GetImages(ctx *gin.Context) {
	methodType := helpers.ParamsDefault(ctx, "methodType", "resize")
	imageSize := helpers.ParamsDefault(ctx, "imageSize", "180x180")
	storageIDEncrypted, _ := ctx.Params.Get("storageID")

	baseEncryption := crypts.NewEncryptionBase().SetUseRandomness(false, os.Getenv("KEY_RANDOM_IMAGE"))
	storageIDDecrypted, err := baseEncryption.Decrypt([]byte(storageIDEncrypted)) // empty passphrase, using default passphrase on env
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("storageID is not valid, error := %s", err.Error()),
		})
		return
	}

	storageID, err := strconv.ParseInt(string(storageIDDecrypted), 10, 64)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("storageID is not found, error := %s", err.Error()),
		})
		return
	}

	var storageModel models.Storages
	err = controller.DB.Where(&models.Storages{Id: storageID}).First(&storageModel).Error
	if err != nil && gorm.IsRecordNotFoundError(err) {
		helpers.NewResponse(ctx, http.StatusNotFound, gin.H{
			"message": fmt.Sprintf("files not found"),
		})
		return
	}

	storageHelpers := storages.NewStorageBase(nil, "", controller.DB)
	file, err := storageHelpers.GetFiles(storageModel)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusNotFound, gin.H{
			"message": fmt.Sprintf("error open file, %s", err.Error()),
		})
		return
	}

	// decode jpeg into image.Image
	img, err := jpeg.Decode(file)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusNotFound, gin.H{
			"message": fmt.Sprintf("error decode file, %s", err.Error()),
		})
		return
	}

	_ = file.Close()

	width, height, isValid, err := storages.GetImageDimensions(imageSize)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("width and height invalid, err := %s", err.Error()),
		})
		return
	} else if !isValid {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("width and height is not supported by the system"),
		})
		return
	}

	var imageResult image.Image
	switch methodType {
	case "resize":
		imageResult = resize.Resize(uint(width), uint(height), img, resize.NearestNeighbor)
	case "thumb":
		imageResult = resize.Thumbnail(uint(width), uint(height), img, resize.NearestNeighbor)
	default:
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"message": fmt.Sprintf("image method is not found, %s", methodType),
		})
		return
	}

	buf := new(bytes.Buffer)
	switch storageModel.Mime {
	case "image/png":
		err = png.Encode(buf, imageResult)
	case "image/gif":
		err = gif.Encode(buf, imageResult, nil)
	case "image/jpeg":
		err = jpeg.Encode(buf, imageResult, nil)
	case "image/jpg":
		err = jpeg.Encode(buf, imageResult, nil)
	default:
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("file is not image"),
		})
		return
	}

	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("image failed to encode, %s", err.Error()),
		})
		return
	}

	responseWriter := ctx.Writer
	responseWriter.Header().Set("Content-Type", storageModel.Mime)
	responseWriter.WriteHeader(http.StatusOK)
	_, _ = io.Copy(responseWriter, buf)
	return
}

func (controller *StorageController) DetailAction(ctx *gin.Context) {
	storageID, _ := ctx.Params.Get("storageID")
	var storageModel models.Storages
	err := controller.DB.Where(&models.Storages{Id: cast.ToInt64(storageID)}).First(&storageModel).Error
	if err != nil && gorm.IsRecordNotFoundError(err) {
		helpers.NewResponse(ctx, http.StatusNotFound, gin.H{
			"message": fmt.Sprintf("files not found"),
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, storageModel)
	return
}
