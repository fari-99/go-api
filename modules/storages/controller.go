package storages

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"

	"github.com/fari-99/go-helper/crypts"
	"github.com/fari-99/go-helper/storages"
	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"

	"go-api/helpers"
)

type controller struct {
	service Service
}

func (c controller) DetailAction(ctx *gin.Context) {
	storageID, isExist := ctx.Params.Get("storageID")
	if !isExist {
		helpers.NewResponse(ctx, http.StatusOK, gin.H{
			"message": "storage id not found",
		})
		return
	}

	storageModel, notFound, err := c.service.GetDetail(ctx, storageID)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusOK, gin.H{
			"error":         err.Error(),
			"error_message": "error getting storage data",
		})
		return
	} else if !notFound {
		helpers.NewResponse(ctx, http.StatusOK, gin.H{
			"error_message": "storage id not found",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, storageModel)
	return
}

func (c controller) S3Policy(ctx *gin.Context) {
	helpers.NewResponse(ctx, http.StatusBadRequest, "nice")
	return
}

func (c controller) UploadAction(ctx *gin.Context) {
	err := ctx.Request.ParseMultipartForm(8 << 20) // 8 MB
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed parsing multipart form, please try again",
		})
		return
	}

	form := ctx.Request.MultipartForm
	storageModels, err := c.service.Uploads(ctx, form)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed upload files, please try again",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, gin.H{
		"message":  "success upload files",
		"storages": storageModels,
	})
	return
}

func (c controller) GetImages(ctx *gin.Context) {
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

	storageModel, notFound, err := c.service.GetDetail(ctx, string(storageIDDecrypted))
	if notFound {
		helpers.NewResponse(ctx, http.StatusNotFound, gin.H{
			"error_message": "file not found",
		})
		return
	} else if err != nil {
		helpers.NewResponse(ctx, http.StatusNotFound, gin.H{
			"error":         err.Error(),
			"error_message": "error get detail storage",
		})
		return
	}

	storageHelpers := storages.NewStorageBase(nil, "")
	file, err := storageHelpers.GetFiles(storageModel.Type, storageModel.Path, storageModel.Filename)
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
