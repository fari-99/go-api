package test_controllers

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"go-api/configs"
	"go-api/helpers"
	"io"
	"net/http"
	"os"

	"github.com/jinzhu/gorm"
)

type FtpController struct {
	DB *gorm.DB
}

func (controller *FtpController) SendFtpAction(ctx *gin.Context) {
	err := ctx.Request.ParseMultipartForm(8 << 20) // 8 MB
	if err != nil {
		configs.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	helpersFtp := helpers.BaseHelperFtp(true).SetCredential(helpers.FtpCredential{})

	form := ctx.Request.MultipartForm
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
		configs.NewResponse(ctx, http.StatusInternalServerError, listError)
		return
	}

	configs.NewResponse(ctx, http.StatusOK, "success send file ftp")
	return
}
