package storages

import (
	"log"

	"github.com/gin-gonic/gin"
)

func NewRegistrator(app *gin.RouterGroup, service Service, authHandler gin.HandlerFunc) {
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
	}
}
