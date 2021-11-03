package copy_modules

import (
	"github.com/gin-gonic/gin"
	"log"
)

func NewRegistrator(app *gin.RouterGroup, service Service, authHandler gin.HandlerFunc) {
	log.Println("Setup Test router")
	control := controller{service: service}

	test := app.Group("/test")
	{
		test.Use(authHandler)
		test.POST("/", control.CreateAction)
		test.GET("/", control.GetListAction)
		test.GET("/{:id}", control.GetDetailAction)
		test.PUT("/{:id}", control.UpdateAction)
		test.DELETE("/{:id}", control.DeleteAction)
	}
}
