package state_machine

import (
	"github.com/gin-gonic/gin"
	"log"
)

func NewRegistrator(app *gin.RouterGroup, service Service, authHandler gin.HandlerFunc) {
	log.Println("Setup State Machine router")
	control := controller{service: service}

	// Approver Endpoint collection
	privateStateMachine := app.Group("/state-machine")
	{
		privateStateMachine.Use(authHandler)
		privateStateMachine.POST("/get-state", control.GetStateTransactionAction)
		privateStateMachine.POST("/change-state", control.ChangeStateAction)
	}
}
