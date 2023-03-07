package calendar_managements

import (
	"log"

	"github.com/gin-gonic/gin"
)

func NewRegistrator(app *gin.RouterGroup, service Service, authHandler gin.HandlerFunc) {
	log.Println("Setup Holiday router")
	control := controller{service: service}

	private := app.Group("/calendar-managements")
	{
		private.Use(authHandler)
		private.POST("/", control.CreateAction)
		private.PUT("/:id", control.UpdateAction)
		private.DELETE("/:id", control.DeleteAction)
	}

	public := app.Group("/calendar-managements")
	{
		public.GET("/", control.GetListAction)
		public.GET("/:id", control.GetDetailAction)
		public.POST("/business-day", control.GetBusinessDayAction)
	}
}
