package test_controllers

import (
	"bytes"
	"go-api/configs"
	"go-api/helpers"
	"io"
	"os"

	"github.com/jinzhu/gorm"
	"github.com/kataras/iris/v12"
)

type FtpController struct {
	DB *gorm.DB
}

func (controller *FtpController) SendFtpAction(ctx iris.Context) {
	// Get the max post value size passed via iris.WithPostMaxMemory.
	maxSize := ctx.Application().ConfigurationReadOnly().GetPostMaxMemory()

	err := ctx.Request().ParseMultipartForm(maxSize)
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusBadRequest, err.Error())
		return
	}

	helpersFtp := helpers.BaseHelperFtp(true).SetCredential(helpers.FtpCredential{})

	form := ctx.Request().MultipartForm
	files := form.File["files[]"]
	var listError []string
	for _, file := range files {
		openFile, _ := file.Open()

		newBuffer := bytes.NewBuffer(nil)
		_, err = io.Copy(newBuffer, openFile)
		if err != nil {
			listError = append(listError, err.Error())
			continue
		}

		err = helpersFtp.
			SetFtpFile(os.Getenv("FTP_TEST_LOCATION"), file.Filename).
			SendFile(newBuffer)
		if err != nil {
			listError = append(listError, err.Error())
			continue
		}

		_ = openFile.Close() // close the file after open it
	}

	if len(listError) > 0 {
		_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, listError)
		return
	}

	_, _ = configs.NewResponse(ctx, iris.StatusOK, "success send file ftp")
	return
}
