package permissions

import (
	"log"

	"github.com/gin-gonic/gin"
)

func NewRegistrator(app *gin.RouterGroup, service Service, authHandler gin.HandlerFunc) {
	log.Println("Setup Permissions RBAC router")
	control := controller{service: service}

	// every route here is gated: management routes must never be reachable without auth + RBAC
	private := app.Group("/permissions")
	{
		private.Use(authHandler)
		private.GET("/", control.GetAction)
		private.POST("/create", control.CreateAction)
		private.DELETE("/delete", control.DeleteAction)
		private.PUT("/update", control.EditAction)
		private.POST("/check", control.CheckAction)
	}
}
