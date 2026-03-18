package security_cameras

import (
	"log"

	"github.com/gin-gonic/gin"
)

func NewRegistrator(app *gin.RouterGroup, service Service, authHandler gin.HandlerFunc) {
	log.Println("Setup Security Cameras router")
	control := controller{service: service}

	test := app.Group("/security-cameras")
	{
		test.Use(authHandler)
		test.POST("/", control.CreateAction)
		test.GET("/", control.GetListAction)
		test.GET("/:id", control.GetDetailAction)
		test.PUT("/:id", control.UpdateAction)
		test.DELETE("/:id", control.DeleteAction)
	}
}
