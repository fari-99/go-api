package controllers

import (
	"go-api/configs"

	"github.com/jinzhu/gorm"
	"github.com/kataras/iris"
)

type StateMachineController struct {
	DB *gorm.DB
}

func (controller *StateMachineController) GetStateTransactionAction(ctx iris.Context) {
	_, _ = configs.NewResponse(ctx, iris.StatusOK, "Yey")
	return
}

func (controller *StateMachineController) ChangeStateAction(ctx iris.Context) {
	_, _ = configs.NewResponse(ctx, iris.StatusOK, "Yey")
	return
}
