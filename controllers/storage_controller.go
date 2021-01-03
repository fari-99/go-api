package controllers

import (
	"bytes"
	"fmt"
	"go-api/configs"
	"go-api/helpers"
	"go-api/models"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"strconv"

	"github.com/jinzhu/gorm"
	"github.com/kataras/iris/v12"
	"github.com/nfnt/resize"
)

type StorageController struct {
	DB *gorm.DB
}

func (controller *StorageController) UploadAction(ctx iris.Context) {
	// Get the max post value size passed via iris.WithPostMaxMemory.
	maxSize := ctx.Application().ConfigurationReadOnly().GetPostMaxMemory()

	err := ctx.Request().ParseMultipartForm(maxSize)
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusBadRequest, err.Error())
		return
	}

	form := ctx.Request().MultipartForm
	files := form.File["files[]"]
	for _, file := range files {
		err = helpers.NewStorageBase(file, "test", controller.DB).SetCtx(ctx).UploadFiles()
		if err != nil {
			_, _ = configs.NewResponse(ctx, iris.StatusBadRequest, err.Error())
			return
		}
	}

	_, _ = configs.NewResponse(ctx, iris.StatusOK, "Success upload files")
	return
}

func (controller *StorageController) GetImages(ctx iris.Context) {
	methodType := ctx.Params().GetStringDefault("methodType", "resize")
	imageSize := ctx.Params().GetStringDefault("imageSize", "180x180")
	storageIDEncrypted := ctx.Params().Get("storageID")

	baseEncryption := helpers.NewEncryptionBase().SetUseRandomness(false, os.Getenv("KEY_RANDOM_IMAGE"))
	storageIDDecrypted, err := baseEncryption.Decrypt([]byte(storageIDEncrypted)) // empty passphrase, using default passphrase on env
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusBadRequest, iris.Map{
			"message": fmt.Sprintf("storageID is not valid, error := %s", err.Error()),
		})
		return
	}

	storageID, err := strconv.ParseInt(string(storageIDDecrypted), 10, 64)
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusBadRequest, iris.Map{
			"message": fmt.Sprintf("storageID is not found, error := %s", err.Error()),
		})
		return
	}

	var storageModel models.Storages
	err = controller.DB.Where(&models.Storages{Id: storageID}).First(&storageModel).Error
	if err != nil && gorm.IsRecordNotFoundError(err) {
		_, _ = configs.NewResponse(ctx, iris.StatusNotFound, iris.Map{
			"message": fmt.Sprintf("files not found"),
		})
		return
	}

	storageHelpers := helpers.NewStorageBase(nil, "", controller.DB)
	file, err := storageHelpers.GetFiles(storageModel)
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusNotFound, iris.Map{
			"message": fmt.Sprintf("error open file, %s", err.Error()),
		})
		return
	}

	// decode jpeg into image.Image
	img, err := jpeg.Decode(file)
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusNotFound, iris.Map{
			"message": fmt.Sprintf("error decode file, %s", err.Error()),
		})
		return
	}

	_ = file.Close()

	width, height, isValid, err := helpers.GetImageDimensions(imageSize)
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusBadRequest, iris.Map{
			"message": fmt.Sprintf("width and height invalid, err := %s", err.Error()),
		})
		return
	} else if !isValid {
		_, _ = configs.NewResponse(ctx, iris.StatusBadRequest, iris.Map{
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
		_, _ = configs.NewResponse(ctx, iris.StatusBadRequest, iris.Map{
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
		_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, iris.Map{
			"message": fmt.Sprintf("file is not image"),
		})
		return
	}

	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, iris.Map{
			"message": fmt.Sprintf("image failed to encode, %s", err.Error()),
		})
		return
	}

	_, _ = configs.NewResponse(ctx, iris.StatusOK, "yey")
	return
}

func (controller *StorageController) DetailAction(ctx iris.Context) {
	_, _ = configs.NewResponse(ctx, iris.StatusOK, "yey")
	return
}
