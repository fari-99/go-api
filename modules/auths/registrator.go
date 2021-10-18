package auths

import (
	"github.com/gin-gonic/gin"
	"log"
)

func NewRegistrator(app *gin.RouterGroup, service Service) {
	log.Println("Setup Auth router")
	control := controller{service: service}

	userPublic := app.Group("/users")
	{
		// authentication data
		userPublic.POST("/auth", control.AuthenticateAction)
	}
}
