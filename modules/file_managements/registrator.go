package file_managements

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// Enabled reports FILE_MANAGEMENT_ENABLED (default false).
func Enabled() bool {
	enabled, _ := strconv.ParseBool(strings.TrimSpace(os.Getenv("FILE_MANAGEMENT_ENABLED")))
	return enabled
}

func NewRegistrator(app *gin.RouterGroup, service Service, authHandler gin.HandlerFunc) {
	log.Println("Setup File Management router")
	control := controller{service: service}

	privateFileManagement := app.Group("/file-managements")
	{
		privateFileManagement.Use(authHandler)
		privateFileManagement.GET("/actions", control.ActionsAction)
		privateFileManagement.GET("", control.ListAction)
		privateFileManagement.POST("", control.CreateAction)
		privateFileManagement.GET("/:fileManagementID", control.DetailAction)
		privateFileManagement.POST("/:fileManagementID/cancel", control.CancelAction)
		privateFileManagement.POST("/:fileManagementID/retry", control.RetryAction)
		privateFileManagement.DELETE("/:fileManagementID", control.DeleteAction)
	}
}
