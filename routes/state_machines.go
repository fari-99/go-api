package routes

import (
	"go-api/controllers"
	"go-api/middleware"
	"log"

	"github.com/kataras/iris/v12"
)

func (routes *Routes) setupStateMachineRoute() *iris.Application {
	log.Println("Setup Customer router")

	app := routes.irisApp
	db := routes.DB

	authentication := middleware.NewMiddleware(middleware.BaseMiddleware{})

	// Approver Endpoint collection
	app.PartyFunc("/state-machine", func(customers iris.Party) {
		stateMachineController := &controllers.StateMachineController{
			DB: db,
		}

		customers.Post("/get-state", stateMachineController.GetStateTransactionAction)
		customers.Post("/change-state", authentication, stateMachineController.ChangeStateAction)
	})

	return app
}
