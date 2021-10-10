package routes

import (
	"go-api/controllers"
	"go-api/middleware"
	"log"
)

func (routes *Routes) setupUserRoute() {
	log.Println("Setup User router")

	app := routes.ginApp
	db := routes.DB
	redis := routes.Redis

	authentication := middleware.AuthMiddleware(middleware.BaseMiddleware{})
	userController := &controllers.UserController{
		DB:    db,
		Redis: redis,
	}

	// Customer Endpoint collection
	usersPublic := app.Group("/customers")
	{
		// authentication data
		usersPublic.POST("/auth", userController.AuthenticateAction)
	}

	usersPrivate := app.Group("/customers")
	{
		usersPrivate.Use(authentication)
		usersPrivate.GET("/details", userController.UserDetailsAction)
		usersPrivate.POST("/create", userController.CreateAction)

		usersPrivate2FA := usersPrivate.Group("/2fa")
		{
			log.Println("Setup User 2FA router")
			users2FAController := &controllers.TwoFactorAuthController{
				DB:    db,
				Redis: redis,
			}

			usersPrivate2FA.POST("/create", users2FAController.CreateNewAuth)
			usersPrivate2FA.POST("/validate", users2FAController.ValidateAuth)
			usersPrivate2FA.GET("/recovery-code", users2FAController.GenerateRecoveryCode)
		}
	}
}
