package routes

import (
	"go-api/controllers"
	"log"

	"github.com/kataras/iris"
)

func (routes *Routes) setupStorageRoute() *iris.Application {
	log.Println("Setup Storage router")

	app := routes.irisApp
	db := routes.DB

	// Approver Endpoint collection
	app.PartyFunc("/storages", func(storages iris.Party) {
		storageController := &controllers.StorageController{DB: db}

		storages.Get("/{:id}", storageController.DetailAction)
		storages.Post("/upload", storageController.UploadAction)
	})

	return app
}
