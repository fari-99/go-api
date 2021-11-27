package twoFA

import (
	"github.com/gin-gonic/gin"
	"log"
)

func NewRegistrator(app *gin.RouterGroup, service Service, authHandler gin.HandlerFunc) {
	log.Println("Setup Customer 2FA router")
	control := controller{service: service}

	user2FA := app.Group("/users/2fa")
	{
		user2FA.Use(authHandler)
		user2FA.POST("/create", control.CreateNewAuth)
		user2FA.POST("/validate", control.ValidateAuth)
		user2FA.GET("/recovery-code", control.GenerateRecoveryCode)
	}
}
