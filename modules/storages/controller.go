package storages

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"mime"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/fari-99/go-helper/crypts"
	"github.com/fari-99/go-helper/storages"
	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"

	"go-api/helpers"
)

type controller struct {
	service Service
}

// DetailAction godoc
// @Summary      Get storage detail
// @Tags         storages
// @Produce      json
// @Param        storageID  path  string  true  "Storage ID"
// @Success      200  {object}  helpers.Response
// @Failure      400  {object}  helpers.Response
// @Failure      404  {object}  helpers.Response
// @Router       /storages/{storageID} [get]
func (c controller) DetailAction(ctx *gin.Context) {
	storageIDParam, isExist := ctx.Params.Get("storageID")
	if !isExist {
		helpers.NewResponse(ctx, http.StatusOK, gin.H{
			"message": "storage id not found",
		})
		return
	}

	storageID, _ := strconv.ParseInt(storageIDParam, 10, 64)
	storageModel, notFound, err := c.service.GetDetail(ctx, uint64(storageID))
	if err != nil {
		helpers.NewResponse(ctx, http.StatusOK, gin.H{
			"error":         err.Error(),
			"error_message": "error getting storage data",
		})
		return
	} else if notFound {
		helpers.NewResponse(ctx, http.StatusNotFound, gin.H{
			"error_message": "storage id not found",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, storageModel)
	return
}

// ListAction godoc
// @Summary      List storages
// @Tags         storages
// @Produce      json
// @Security     BearerAuth
// @Param        page  query  int  false  "Page (default 1)"
// @Param        limit  query  int  false  "Page size (default 20)"
// @Success      200  {object}  helpers.Response
// @Failure      500  {object}  helpers.Response
// @Router       /storages [get]
func (c controller) ListAction(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	items, paginatorData, err := c.service.GetList(ctx, page, limit)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"error":         err.Error(),
			"error_message": "failed to list files",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, gin.H{
		"paginator": paginatorData,
		"items":     items,
	})
}

// ContentAction streams the file with its stored MIME type. Range requests are supported, so
// audio/video can seek. Only the owner can read it. ?download=1 forces a download.
// ContentAction godoc
// @Summary      Get file content
// @Tags         storages
// @Produce      octet-stream
// @Security     BearerAuth
// @Param        storageID  path  int  true  "Storage ID"
// @Param        download  query  string  false  "1 to force download"
// @Success      200  {file}  file
// @Failure      400  {object}  helpers.Response
// @Failure      404  {object}  helpers.Response
// @Router       /storages/{storageID}/content [get]
func (c controller) ContentAction(ctx *gin.Context) {
	storageID, err := strconv.ParseUint(ctx.Param("storageID"), 10, 64)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{"message": "storageID is not valid"})
		return
	}

	storageModel, file, err := c.service.OpenOwned(ctx, storageID)
	if errors.Is(err, ErrStorageNotFound) {
		helpers.NewResponse(ctx, http.StatusNotFound, gin.H{"message": "file not found"})
		return
	} else if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{"message": "failed to open file"})
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{"message": "failed to read file"})
		return
	}

	// only media is shown inline; anything else (html, text, ...) is always a download
	disposition := "attachment"
	if ctx.Query("download") != "1" && isInlineMime(storageModel.Mime) {
		disposition = "inline"
	}

	name := strings.NewReplacer("\"", "", "\r", "", "\n", "").Replace(storageModel.OriginalFilename)
	ctx.Header("Content-Type", storageModel.Mime)
	ctx.Header("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": name}))
	ctx.Header("X-Content-Type-Options", "nosniff")
	ctx.Header("Content-Security-Policy", "default-src 'none'; sandbox")
	ctx.Header("Cache-Control", "private, max-age=300")
	http.ServeContent(ctx.Writer, ctx.Request, "", stat.ModTime(), file)
}

// S3Policy godoc
// @Summary      Get S3 upload policy
// @Tags         storages
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  helpers.Response
// @Router       /storages/s3-policy [post]
func (c controller) S3Policy(ctx *gin.Context) {
	helpers.NewResponse(ctx, http.StatusBadRequest, "nice")
	return
}

// UploadAction godoc
// @Summary      Upload files (multipart, max 8MB)
// @Tags         storages
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  helpers.Response
// @Failure      400  {object}  helpers.Response
// @Router       /storages/upload [post]
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

// GetImages godoc
// @Summary      Get resized image
// @Tags         storages
// @Produce      image/*
// @Param        storageID  path  string  true  "Encrypted storage ID"
// @Param        methodType  path  string  true  "Resize method (default resize)"
// @Param        imageSize  path  string  true  "WxH, e.g. 180x180"
// @Success      200  {file}  file
// @Failure      400  {object}  helpers.Response
// @Failure      404  {object}  helpers.Response
// @Router       /storages/{storageID}/{methodType}/{imageSize} [get]
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

	storageID, _ := strconv.ParseInt(string(storageIDDecrypted), 10, 64)
	storageModel, notFound, err := c.service.GetDetail(ctx, uint64(storageID))
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

func isInlineMime(mimeType string) bool {
	return strings.HasPrefix(mimeType, "image/") || strings.HasPrefix(mimeType, "video/") || strings.HasPrefix(mimeType, "audio/")
}
