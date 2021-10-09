package routes

import (
	"go-api/controllers"
	"go-api/middleware"
	"log"
)

func (routes *Routes) setupCustomerRoute() {
	log.Println("Setup Customer router")

	app := routes.ginApp
	db := routes.DB
	redis := routes.Redis

	authentication := middleware.AuthMiddleware(middleware.BaseMiddleware{})
	customerController := &controllers.CustomerController{
		DB:    db,
		Redis: redis,
	}

	// Customer Endpoint collection
	customersPublic := app.Group("/customers")
	{
		// authentication data
		customersPublic.POST("/auth", customerController.AuthenticateAction)
	}

	customersPrivate := app.Group("/customers")
	{
		customersPrivate.Use(authentication)
		customersPrivate.GET("/details", customerController.CustomerDetailsAction)
		customersPrivate.POST("/create", customerController.CreateAction)

		customersPrivate2FA := customersPrivate.Group("/2fa")
		{
			log.Println("Setup Customer 2FA router")
			customers2FAController := &controllers.TwoFactorAuthController{
				DB:    db,
				Redis: redis,
			}

			customersPrivate2FA.POST("/create", customers2FAController.CreateNewAuth)
			customersPrivate2FA.POST("/validate", customers2FAController.ValidateAuth)
			customersPrivate2FA.GET("/recovery-code", customers2FAController.GenerateRecoveryCode)
		}
	}
}
