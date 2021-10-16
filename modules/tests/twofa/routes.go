package twofa

import (
	"github.com/gin-gonic/gin"
	"go-api/modules/configs"
	middleware2 "go-api/modules/middleware"
	"log"
)

func NewRoute(app *gin.Engine) {
	log.Println("Setup Test 2FA router")

	twoFactorAuth := app.Group("/test-two-auth")
	{
		twoFactorAuthController := &TwoFactorAuthController{
			DI: configs.DIInit(),
		}

		twoFactorAuth.POST("/new", twoFactorAuthController.NewAuth)
		twoFactorAuth.POST("/validate", twoFactorAuthController.Validate)

		otpMiddleware := middleware2.OTPMiddleware(middleware2.BaseMiddleware{})
		authMiddleware := middleware2.AuthMiddleware(middleware2.BaseMiddleware{})
		twoFactorAuth.Use(authMiddleware, otpMiddleware).GET("/test", twoFactorAuthController.TestMiddleware)
	}
}
