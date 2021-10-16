package finite_states

import (
	"github.com/gin-gonic/gin"
	"go-api/modules/configs"
	"log"
)

func NewRoute(app *gin.Engine) {
	log.Printf("Setup Test Finite State router")

	testStateMachine := app.Group("/test-state-machine")
	{
		stateMachineController := &FiniteStateController{
			DI: configs.DIInit(),
		}

		testStateMachine.POST("/get-state", stateMachineController.GetAvailableTransitionsAction)
		testStateMachine.POST("/change-state", stateMachineController.ChangeStateAction)
	}
}
