package state_machine

import (
	"github.com/gin-gonic/gin"
	"go-api/helpers"
	"go-api/modules/configs"
	"net/http"
)

type StateMachineController struct {
	*configs.DI
}

func (controller *StateMachineController) GetStateTransactionAction(ctx *gin.Context) {
	helpers.NewResponse(ctx, http.StatusOK, "Yey")
	return
}

func (controller *StateMachineController) ChangeStateAction(ctx *gin.Context) {
	helpers.NewResponse(ctx, http.StatusOK, "Yey")
	return
}
