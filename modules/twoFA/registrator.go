package twoFA

import (
	"log"

	"github.com/gin-gonic/gin"
)

func NewRegistrator(app *gin.RouterGroup, service Service, authHandler gin.HandlerFunc) {
	log.Println("Setup Customer 2FA router")
	control := controller{service: service}

	totp2FA := app.Group("/users/2fa/totp")
	{
		totp2FA.Use(authHandler)

		// 2FA
		totp2FA.POST("/create", control.CreateTotp)
		totp2FA.POST("/validate/:action", control.ValidateTotp)
		totp2FA.PUT("/disabled", control.DisabledTotp)
	}

	recoveryCode2FA := app.Group("/users/2fa/recovery-code")
	{
		recoveryCode2FA.Use(authHandler)

		// Recovery Code
		recoveryCode2FA.POST("/create", control.CreateRecoveryCode)
		recoveryCode2FA.POST("/validate/:action", control.ValidateRecoveryCode)
		recoveryCode2FA.PUT("/disabled", control.DisabledTotp) // TODO: add disable recovery code
	}
}
