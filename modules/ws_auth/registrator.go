package ws_auth

import (
	wsauth "github.com/fari-99/go-helper/ws_auth"
	"log"

	"github.com/gin-gonic/gin"
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
