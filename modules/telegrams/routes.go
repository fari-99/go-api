package telegrams

import (
	"github.com/gin-gonic/gin"
	"go-api/modules/configs"
	middleware2 "go-api/modules/middleware"
	"log"
)

func NewRoute(app *gin.Engine) {
	log.Println("Setup Telegram router")

	authentication := middleware2.AuthMiddleware(middleware2.BaseMiddleware{})

	// Approver Endpoint collection
	telegrams := app.Group("/telegrams").Use(authentication)
	{
		telegramController := &TelegramController{
			DI: configs.DIInit(),
		}
		//companyIDPathName := "companyID"

		// authentication data
		telegrams.POST("/authenticate", authentication, telegramController.AuthenticateAction)
	}
}
