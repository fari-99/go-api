package state_machine

import (
	"github.com/gin-gonic/gin"
	"go-api/modules/configs"
	middleware2 "go-api/modules/middleware"
	"log"
)

func NewRoute(app *gin.Engine) {
	log.Println("Setup State Machine router")

	authentication := middleware2.AuthMiddleware(middleware2.BaseMiddleware{})

	// Approver Endpoint collection
	stateMachine := app.Group("/state-machine").Use(authentication)
	{
		stateMachineController := &StateMachineController{
			DI: configs.DIInit(),
		}

		stateMachine.POST("/get-state", stateMachineController.GetStateTransactionAction)
		stateMachine.POST("/change-state", authentication, stateMachineController.ChangeStateAction)
	}
}
