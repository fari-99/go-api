package routes

import (
	"go-api/configs"
	"go-api/controllers"
	"log"

	"github.com/kataras/iris"
)

func (routes *Routes) setupStateMachineRoute() *iris.Application {
	log.Println("Setup Customer router")

	app := routes.irisApp
	db := routes.DB

	authentication := configs.NewMiddleware(configs.MiddlewareConfiguration{})

	// Approver Endpoint collection
	app.PartyFunc("/state-machine", func(customers iris.Party) {
		stateMachineController := &controllers.StateMachineController{DB: db}

		customers.Post("/get-state", stateMachineController.GetStateTransactionAction)
		customers.Post("/change-state", authentication, stateMachineController.ChangeStateAction)
	})

	return app
}
