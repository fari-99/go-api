package emails

import (
	"github.com/gin-gonic/gin"
	"go-api/modules/configs"
	"log"
)

func NewRoute(app *gin.Engine) {
	log.Printf("Setup Test EmailDialler Connection router")

	testEmail := app.Group("/test-email")
	{
		testEmailController := &EmailsController{
			DI: configs.DIInit(),
		}

		testEmail.POST("/send-email", testEmailController.SendEmailAction)
	}
}
