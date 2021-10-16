package ftps

import (
	"github.com/gin-gonic/gin"
	"go-api/modules/configs"
	"log"
)

func NewRoute(app *gin.Engine) {
	log.Println("Setup Test FTP router")

	testFtp := app.Group("/test-ftp")
	{
		ftpController := &FtpController{
			DI: configs.DIInit(),
		}

		testFtp.POST("/send-files", ftpController.SendFtpAction)
		//testFtp.POST("/send-files-location")
		//testFtp.POST("/send-files-open-files")
	}
}
