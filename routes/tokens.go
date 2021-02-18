package routes

import (
	"go-api/controllers"
	"log"
)

func (routes *Routes) setupTokenRoute() {
	log.Println("Setup Token router")

	app := routes.ginApp
	db := routes.DB

	// Approver Endpoint collection
	tokens := app.Group("/token")
	{
		tokenController := &controllers.TokenController{
			DB: db,
		}
		//companyIDPathName := "companyID"

		tokens.POST("/create", tokenController.CreateTokenAction)
		tokens.POST("/check", tokenController.CheckTokenAction)
	}
}
