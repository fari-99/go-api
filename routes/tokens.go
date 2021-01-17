package routes

import (
	"go-api/controllers"
	"log"

	"github.com/kataras/iris/v12"
)

func (routes *Routes) setupTokenRoute() *iris.Application {
	log.Println("Setup Token router")

	app := routes.irisApp
	db := routes.DB

	// Approver Endpoint collection
	app.PartyFunc("/token", func(tokens iris.Party) {
		tokenController := &controllers.TokenController{
			DB: db,
		}
		//companyIDPathName := "companyID"

		tokens.Post("/create", tokenController.CreateTokenAction)
		tokens.Post("/check", tokenController.CheckTokenAction)
	})

	return app
}
