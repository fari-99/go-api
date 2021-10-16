package users

import (
	"github.com/gin-gonic/gin"
	"go-api/modules/configs"
	middleware2 "go-api/modules/middleware"
	"log"
)

func NewRoute(app *gin.Engine) {
	log.Println("Setup User router")

	authentication := middleware2.AuthMiddleware(middleware2.BaseMiddleware{})
	userController := &UserController{
		DI: configs.DIInit(),
	}

	userPublic := app.Group("/customers")
	{
		// authentication data
		userPublic.POST("/auth", userController.AuthenticateAction)
	}

	userPrivate := app.Group("/customers")
	{
		userPrivate.Use(authentication)
		userPrivate.GET("/details", userController.CustomerDetailsAction)
		userPrivate.POST("/create", userController.CreateAction)

		userPrivate2FA := userPrivate.Group("/2fa")
		{
			log.Println("Setup Customer 2FA router")
			user2FAController := &TwoFactorAuthController{
				DI: configs.DIInit(),
			}

			userPrivate2FA.POST("/create", user2FAController.CreateNewAuth)
			userPrivate2FA.POST("/validate", user2FAController.ValidateAuth)
			userPrivate2FA.GET("/recovery-code", user2FAController.GenerateRecoveryCode)
		}
	}
}
