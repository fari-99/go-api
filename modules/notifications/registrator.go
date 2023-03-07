package notifications

import (
	"log"

	"github.com/gin-gonic/gin"
)

func NewRegistrator(app *gin.RouterGroup, service Service, authHandler gin.HandlerFunc) {
	log.Println("Setup Notification router")
	control := controller{service: service}

	notifications := app.Group("/notifications")
	{
		notifications.Use(authHandler)
		notifications.POST("/", control.CreateAction)
		notifications.GET("/", control.GetListAction)
		notifications.GET("/:id", control.GetDetailAction)
		notifications.PUT("/:id", control.UpdateAction)
		notifications.DELETE("/:id", control.DeleteAction)

		notifications.GET("/qr-code/whatsapp", control.GetQRCodeWhatsapp)
	}
}
