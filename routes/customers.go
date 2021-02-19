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
		customersPublic.POST("/registers/:customerType", customerController.RegisterAction)
	}

	customersPrivate := app.Group("/customers").Use(authentication)
	{
		customersPrivate.GET("/details", customerController.CustomerDetailsAction)
		customersPrivate.POST("/create", customerController.CreateAction) //
	}
}
