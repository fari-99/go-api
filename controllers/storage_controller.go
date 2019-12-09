package controllers

import (
	"go-api/configs"
	"go-api/helpers"

	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
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

func (controller *StorageController) DetailAction(ctx iris.Context) {
	_, _ = configs.NewResponse(ctx, iris.StatusOK, "yey")
	return
}
