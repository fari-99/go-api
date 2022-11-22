package ftps

import (
	"bytes"
	"io"
	"net/http"
	"os"

	gohelper "github.com/fari-99/go-helper"
	"github.com/gin-gonic/gin"

	"go-api/helpers"
	"go-api/modules/configs"
)

type FtpController struct {
	*configs.DI
}

func (controller *FtpController) SendFtpAction(ctx *gin.Context) {
	err := ctx.Request.ParseMultipartForm(8 << 20) // 8 MB
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	ftpCredential := gohelper.FtpCredential{
		FtpHost:     os.Getenv("FTP_TEST_HOST"),
		FtpPort:     os.Getenv("FTP_PORT"), // sftp port default 22
		SshUser:     os.Getenv("FTP_TEST_USERNAME"),
		SshPassword: os.Getenv("FTP_TEST_PASSWORD"),
		SshKeyFile:  os.Getenv("FTP_AUTH_FILE_LOCATION") + os.Getenv("FTP_TEST_AUTH_FILE"),
		FtpUser:     "",
		FtpPassword: "",
	}

	helpersFtp := gohelper.BaseHelperFtp(true).SetCredential(ftpCredential)

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
		helpers.NewResponse(ctx, http.StatusInternalServerError, listError)
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "success send file ftp")
	return
}
