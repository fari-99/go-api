package roles

import (
	"log"

	"github.com/gin-gonic/gin"
)

func NewRegistrator(app *gin.RouterGroup, service Service, authHandler gin.HandlerFunc) {
	log.Println("Setup Roles router")
	control := controller{service: service}

	rolesGroup := app.Group("/roles")
	{
		rolesGroup.Use(authHandler)
		rolesGroup.POST("/", control.CreateAction)
		rolesGroup.GET("/", control.GetListAction)
		rolesGroup.GET("/:id", control.GetDetailAction)
		rolesGroup.GET("/:id/permissions-count", control.CountPermissionsAction)
		rolesGroup.PUT("/:id", control.UpdateAction)
		rolesGroup.DELETE("/:id", control.DeleteAction)
	}
}
