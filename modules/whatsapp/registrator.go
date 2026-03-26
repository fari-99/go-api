package whatsapp

import (
	"log"

	"go-api/modules/configs"

	"github.com/gin-gonic/gin"
)

func NewRegistrator(app *gin.RouterGroup, di *configs.DI, authHandler gin.HandlerFunc) {
	log.Println("Setup Whatsapp client router")
	control := controller{di: di}

	whatsappClient := app.Group("/whatsapp")
	{
		whatsappClient.Use(authHandler)
		whatsappClient.POST("/login", control.LoginAction)
		whatsappClient.POST("/qr-code", control.QRCodeAction)
	}
}
