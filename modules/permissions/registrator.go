package permissions

import (
	"log"

	"github.com/gin-gonic/gin"
)

func NewRegistrator(app *gin.RouterGroup, service Service, authHandler gin.HandlerFunc) {
	log.Println("Setup Permissions RBAC router")
	control := controller{service: service}

	// Companies routes
	public := app.Group("/permissions")
	{
		public.GET("/", control.GetAction)
		public.POST("/create", control.CreateAction)
		public.DELETE("/delete", control.DeleteAction)
		public.PUT(`/update`, control.EditAction)
	}

	private := app.Group("/permissions")
	{
		private.Use(authHandler)
		private.POST("/check", control.CheckAction)
	}
}
