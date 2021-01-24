package routes

import (
	"go-api/controllers"
	"go-api/middleware"
	"log"

	"github.com/kataras/iris/v12"
)

func (routes *Routes) setupStorageRoute() *iris.Application {
	log.Println("Setup Storage router")

	app := routes.irisApp
	db := routes.DB

	authentication := middleware.NewMiddleware(middleware.BaseMiddleware{})

	// Storages Endpoint collection
	app.PartyFunc("/storages", func(storages iris.Party) {
		storageController := &controllers.StorageController{
			DB: db,
		}

		storages.Get("/{:id}", authentication, storageController.DetailAction)
		storages.Post("/upload", authentication, storageController.UploadAction)
	})

	return app
}
