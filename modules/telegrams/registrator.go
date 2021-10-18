package telegrams

import (
	"github.com/gin-gonic/gin"
	"log"
)

func NewRegistrator(app *gin.RouterGroup, service Service, authHandler gin.HandlerFunc) {
	log.Println("Setup Telegram router")
	control := controller{service: service}

	telegrams := app.Group("/telegrams")
	{
		telegrams.Use(authHandler)
		telegrams.POST("/authenticate", control.AuthenticateAction)
	}
}
