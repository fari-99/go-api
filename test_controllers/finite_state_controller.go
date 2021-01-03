package test_controllers

import (
	"encoding/json"
	"fmt"
	"go-api/bussiness_flow"
	"go-api/configs"
	"go-api/models"

	"github.com/jinzhu/gorm"
	"github.com/kataras/iris/v12"
)

type FiniteStateController struct {
	DB *gorm.DB
}

type InputFSM struct {
	ConfigName     string `json:"config_name"`
	TransitionName string `json:"transition_name"`
	IsFiltered     bool   `json:"is_filtered"`
	DataID         int64  `json:"data_id"`
}

func (controller *FiniteStateController) GetAvailableTransitionsAction(ctx iris.Context) {
	var input InputFSM
	err := ctx.ReadJSON(&input)
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusBadRequest, err.Error())
		return
	}

	var transactionModel models.Transactions
	err = controller.DB.First(&transactionModel, input.DataID).Error
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusNotFound, err.Error())
		return
	}

	var dataTransitions map[string]interface{}
	dataMarshal, _ := json.Marshal(transactionModel)
	_ = json.Unmarshal(dataMarshal, &dataTransitions)

	// if some times in the future filtered available must be from user input then change false to filteredAvailable
	baseSM, err := bussiness_flow.NewBaseStateMachine(
		input.ConfigName,
		input.IsFiltered,
		input.TransitionName,
		dataTransitions)
	if err != nil {
		msg := fmt.Errorf("error create base, err := %s", err.Error())
		_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, msg.Error())
		return
	}

	// Get properties for current state
	currentProperties, err := baseSM.GetStateProperties()
	if err != nil {
		msg := fmt.Errorf("error get current properties, err := %s", err.Error())
		_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, msg.Error())
		return
	}

	// Get available transitions
	availableTransitions, err := baseSM.GetAvailableTransitions()
	if err != nil {
		msg := fmt.Errorf("error get available transition, err := %s", err.Error())
		_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, msg.Error())
		return
	}

	_, _ = configs.NewResponse(ctx, iris.StatusOK, iris.Map{
		"currentProperties":    currentProperties,
		"availableTransitions": availableTransitions,
	})
	return
}

func (controller *FiniteStateController) ChangeStateAction(ctx iris.Context) {
	var input InputFSM
	err := ctx.ReadJSON(&input)
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusBadRequest, err.Error())
		return
	}

	var transactionModel models.Transactions
	err = controller.DB.First(&transactionModel, input.DataID).Error
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusNotFound, err.Error())
		return
	}

	var dataTransitions map[string]interface{}
	dataMarshal, _ := json.Marshal(transactionModel)
	_ = json.Unmarshal(dataMarshal, &dataTransitions)

	// create State Machine Base
	baseSM, err := bussiness_flow.NewBaseStateMachine(
		input.ConfigName,
		input.IsFiltered,
		input.TransitionName,
		dataTransitions)
	if err != nil {
		msg := fmt.Errorf("error create base, err := %s", err.Error())
		_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, msg.Error())
		return
	}

	// Change State
	isChanged, err := baseSM.ChangeStateMachine()
	if err != nil {
		msg := fmt.Errorf("state not changed, err := %s", err.Error())
		_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, msg.Error())
		return
	}

	constraitValue, err := baseSM.GetStatusByName(baseSM.Fsm.Current())
	if err != nil {
		msg := fmt.Errorf("can't get status by name, err := %s", err.Error())
		_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, msg.Error())
		return
	}

	baseSM.CurrentStatus = constraitValue

	// Get properties for current state
	currentProperties, err := baseSM.GetStateProperties()
	if err != nil {
		msg := fmt.Errorf("can't get state properties, err := %s", err.Error())
		_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, msg.Error())
		return
	}

	_, _ = configs.NewResponse(ctx, iris.StatusOK, iris.Map{
		"is_changed":           isChanged,
		"transition_used_name": input.TransitionName,
		"new_state_value":      constraitValue,
		"state_properties":     currentProperties,
	})
	return
}
