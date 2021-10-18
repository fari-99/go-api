package twofa

import (
	"github.com/gin-gonic/gin"
	"go-api/modules/configs"
	"go-api/modules/middleware"
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

		otpMiddleware := middleware.OTPMiddleware(middleware.BaseMiddleware{})
		authMiddleware := middleware.AuthMiddleware(middleware.BaseMiddleware{})
		twoFactorAuth.Use(authMiddleware, otpMiddleware).GET("/test", twoFactorAuthController.TestMiddleware)
	}
}
