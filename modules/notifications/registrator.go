package notifications

import (
	"log"

	"github.com/gin-gonic/gin"
)

func NewRegistrator(app *gin.RouterGroup, service Service, authHandler gin.HandlerFunc) {
	log.Println("Setup Notification router")
	control := controller{service: service}

	test := app.Group("/notifications")
	{
		test.Use(authHandler)
		test.POST("/", control.CreateAction)
		test.GET("/", control.GetListAction)
		test.GET("/:id", control.GetDetailAction)
		test.PUT("/:id", control.UpdateAction)
		test.DELETE("/:id", control.DeleteAction)
	}
}
