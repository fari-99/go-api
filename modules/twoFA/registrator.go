package twoFA

import (
	"log"

	"github.com/gin-gonic/gin"
)

func NewRegistrator(app *gin.RouterGroup, service Service, authHandler gin.HandlerFunc) {
	log.Println("Setup Customer 2FA router")
	control := controller{service: service}

	user2FA := app.Group("/users/2fa")
	{
		user2FA.Use(authHandler)

		// 2FA
		user2FA.POST("/create", control.CreateNewAuth)
		user2FA.POST("/validate", control.ValidateAuth)
		user2FA.PUT("/disabled", control.DisabledAuth)
	}

	userRecoveryCode := app.Group("/users/recovery-code")
	{
		userRecoveryCode.Use(authHandler)

		// Recovery Code
		userRecoveryCode.GET("/create", control.GenerateRecoveryCode)
		userRecoveryCode.POST("/validate", control.ValidateRecoveryCodeAuth)
		userRecoveryCode.PUT("/disabled", control.DisabledAuth) // TODO: add disable recovery code
	}
}
