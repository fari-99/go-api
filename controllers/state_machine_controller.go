package controllers

import (
	"github.com/gin-gonic/gin"
	"go-api/configs"
	"net/http"

	"github.com/jinzhu/gorm"
)

type StateMachineController struct {
	DB *gorm.DB
}

func (controller *StateMachineController) GetStateTransactionAction(ctx *gin.Context) {
	configs.NewResponse(ctx, http.StatusOK, "Yey")
	return
}

func (controller *StateMachineController) ChangeStateAction(ctx *gin.Context) {
	configs.NewResponse(ctx, http.StatusOK, "Yey")
	return
}
