package storages

import (
	"github.com/gin-gonic/gin"
	"go-api/modules/configs"
	middleware2 "go-api/modules/middleware"
	"log"
)

func NewRoute(app *gin.Engine) {
	log.Println("Setup Storage router")

	authentication := middleware2.AuthMiddleware(middleware2.BaseMiddleware{})

	// Storages Endpoint collection
	storages := app.Group("/storages").Use(authentication)
	{
		storageController := &StorageController{
			DI: configs.DIInit(),
		}

		storages.GET("/{:id}", authentication, storageController.DetailAction)
		storages.POST("/upload", authentication, storageController.UploadAction)
	}
}
