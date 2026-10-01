package storages

import (
	"log"

	"github.com/gin-gonic/gin"

	wsauth "github.com/fari-99/go-helper/ws_auth"
	"go-api/modules/ws_auth"
)

func NewRegistrator(app *gin.RouterGroup, service Service, authHandler gin.HandlerFunc, wsAuth wsauth.Service) {
	log.Println("Setup Storage router")
	control := controller{service: service}

	// Storages Endpoint collection
	publicStorage := app.Group("/storages")
	{
		publicStorage.GET("/:storageID", control.DetailAction)
		publicStorage.GET("/:storageID/:methodType/:imageSize", control.GetImages)
	}

	privateStorage := app.Group("/storages")
	{
		privateStorage.POST("/upload", authHandler, control.UploadAction)
		privateStorage.POST("/s3-policy", authHandler, control.S3Policy)
	}

	// WebSocket upload, authenticated with the short-lived encrypted WS token (see modules/ws_auth)
	wsControl := newWSController(control.service, wsAuth)
	app.GET("/ws/storages/upload", ws_auth.Middleware(wsAuth), wsControl.UploadAction)
}
