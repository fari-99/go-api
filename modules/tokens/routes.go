package tokens

import (
	"github.com/gin-gonic/gin"
	"go-api/modules/configs"
	"log"
)

func NewRoute(app *gin.Engine) {
	log.Println("Setup Token router")

	// Token Endpoint collection
	tokens := app.Group("/token")
	{
		tokenController := &TokenController{
			DI: configs.DIInit(),
		}

		tokens.POST("/create", tokenController.CreateTokenAction)
		tokens.POST("/check", tokenController.CheckTokenAction)
	}
}
