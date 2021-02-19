package routes

import (
	"go-api/controllers"
	"go-api/middleware"
	"log"
)

func (routes *Routes) setupStateMachineRoute() {
	log.Println("Setup State Machine router")

	app := routes.ginApp
	db := routes.DB

	authentication := middleware.AuthMiddleware(middleware.BaseMiddleware{})

	// Approver Endpoint collection
	stateMachine := app.Group("/state-machine").Use(authentication)
	{
		stateMachineController := &controllers.StateMachineController{
			DB: db,
		}

		stateMachine.POST("/get-state", stateMachineController.GetStateTransactionAction)
		stateMachine.POST("/change-state", stateMachineController.ChangeStateAction)
	}
}
