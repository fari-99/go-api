package ws_auth

import (
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"

	wsauth "github.com/fari-99/go-helper/ws_auth"
)

func NewRegistrator(app *gin.RouterGroup, service wsauth.Service, authHandler gin.HandlerFunc) {
	log.Println("Setup WS Auth router")
	control := controller{service: service}

	tokens := app.Group("/storages/ws-token")
	{
		tokens.POST("", authHandler, control.IssueAction)
		tokens.POST("/refresh", control.RefreshAction)
	}
}

// Enabled reports whether WS_UPLOAD_ENABLED is true. Defaults to false.
func Enabled() bool {
	enabled, _ := strconv.ParseBool(os.Getenv("WS_UPLOAD_ENABLED"))
	return enabled
}
