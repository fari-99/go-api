package state_machine

import (
	"github.com/gin-gonic/gin"
	"go-api/helpers"
	"net/http"
)

type controller struct {
	service Service
}

// GetStateTransactionAction godoc
// @Summary      Get transaction state
// @Tags         state-machine
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  helpers.Response
// @Router       /state-machine/get-state [post]
func (c controller) GetStateTransactionAction(ctx *gin.Context) {
	helpers.NewResponse(ctx, http.StatusOK, "Yey")
	return
}

// ChangeStateAction godoc
// @Summary      Change transaction state
// @Tags         state-machine
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  helpers.Response
// @Router       /state-machine/change-state [post]
func (c controller) ChangeStateAction(ctx *gin.Context) {
	helpers.NewResponse(ctx, http.StatusOK, "Yey")
	return
}
