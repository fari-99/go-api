package locations

import (
	"log"

	"github.com/gin-gonic/gin"
)

func NewRegistrator(app *gin.RouterGroup, service Service, authHandler gin.HandlerFunc) {
	log.Println("Setup Locations router")
	control := controller{service: service}

	locationPrivate := app.Group("/locations")
	{
		locationPrivate.Use(authHandler)
		locationPrivate.POST("/create", control.CreateAction)
		locationPrivate.PUT("/:locationID", control.UpdateAction)
		locationPrivate.DELETE("/:locationID", control.DeleteAction)
		locationPrivate.PUT("/:locationID/status/:status", control.UpdateStatusAction)
	}

	locationPublic := app.Group("/locations")
	{
		locationPublic.GET("/", control.GetAllAction)               // get all locations
		locationPublic.GET("/:locationID", control.GetDetailAction) // get all location by id
	}

	locationLevelsPrivate := app.Group("/locations/levels")
	{
		locationLevelsPrivate.Use(authHandler)
		locationLevelsPrivate.GET("/", control.GetAllActionLevel)
		locationLevelsPrivate.GET("/:levelID", control.GetDetailActionLevel)
		locationLevelsPrivate.POST("/create", control.CreateActionLevel)
		locationLevelsPrivate.PUT("/:levelID", control.UpdateActionLevel)
		locationLevelsPrivate.PUT("/:levelID/status/:status", control.UpdateStatusActionLevel)
		locationLevelsPrivate.DELETE("/:levelID/delete", control.DeleteActionLevel)
	}
}
