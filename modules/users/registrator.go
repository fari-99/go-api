package users

import (
	"github.com/gin-gonic/gin"
	"log"
)

func NewRegistrator(app *gin.RouterGroup, service Service, authHandler gin.HandlerFunc) {
	log.Println("Setup User router")
	control := controller{service: service}

	userPrivate := app.Group("/users")
	{
		userPrivate.Use(authHandler)
		userPrivate.POST("/create", control.CreateAction)
	}
}
