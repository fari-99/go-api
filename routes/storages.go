package routes

import (
	"go-api/controllers"
	"go-api/middleware"
	"log"
)

func (routes *Routes) setupStorageRoute() {
	log.Println("Setup Storage router")

	app := routes.ginApp
	db := routes.DB

	authentication := middleware.AuthMiddleware(middleware.BaseMiddleware{})

	// Storages Endpoint collection
	storages := app.Group("/storages").Use(authentication)
	{
		storageController := &controllers.StorageController{
			DB: db,
		}

		storages.GET("/{:id}", authentication, storageController.DetailAction)
		storages.POST("/upload", authentication, storageController.UploadAction)
	}
}
