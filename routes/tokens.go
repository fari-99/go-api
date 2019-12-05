package routes

import (
	"go-api/configs"
	"go-api/controllers"
	"log"

	"github.com/kataras/iris"
)

func (routes *Routes) setupTokenRoute() *iris.Application {
	log.Println("Setup Token router")

	app := routes.irisApp
	db := routes.DB

	authentication := configs.NewMiddleware(configs.MiddlewareConfiguration{})

	// Approver Endpoint collection
	app.PartyFunc("/token", func(tokens iris.Party) {
		tokenController := &controllers.TokenController{DB: db}
		//companyIDPathName := "companyID"

		tokens.Post("/create", tokenController.CreateTokenAction)
		tokens.Post("/check", authentication, tokenController.CheckTokenAction)
	})

	return app
}
