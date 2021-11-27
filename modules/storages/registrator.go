package storages

import (
	"github.com/gin-gonic/gin"
	"log"
)

func NewRegistrator(app *gin.RouterGroup, service Service, authHandler gin.HandlerFunc) {
	log.Println("Setup Storage router")
	control := controller{service: service}

	// Storages Endpoint collection
	publicStorage := app.Group("/storages")
	{
		publicStorage.GET("/{:id}", control.DetailAction)
		publicStorage.GET("/{:id}/{:methodType}/{:imageSize}", control.GetImages)
	}

	privateStorage := app.Group("/storages")
	{
		privateStorage.POST("/upload", authHandler, control.UploadAction)
	}
}
