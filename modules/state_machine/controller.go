package state_machine

import (
	"github.com/gin-gonic/gin"
	"go-api/helpers"
	"net/http"
)

type controller struct {
	service Service
}

func (c controller) GetStateTransactionAction(ctx *gin.Context) {
	helpers.NewResponse(ctx, http.StatusOK, "Yey")
	return
}

func (c controller) ChangeStateAction(ctx *gin.Context) {
	helpers.NewResponse(ctx, http.StatusOK, "Yey")
	return
}
