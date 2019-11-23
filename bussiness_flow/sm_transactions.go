package bussiness_flow

import (
	"fmt"
	"go-api/bussiness_flow/properties/emails"
	"go-api/bussiness_flow/properties/new_objects"
	"go-api/constant"
	"go-api/constant/constant_models"

	"github.com/looplab/fsm"
)

// ==================== Constanta Model PO to ConstantBase for FSM =====================

/**
Note : every time you create new finite state file, you must add your configName to base.go on this function
	- GetAvailableTransitions
	- getEventState
	- getAllStatus
*/
func (base *BaseSMTransaction) GetTransactionStatus() (map[int]*ConstantBase, error) {
	// init Statuses
	var PoCStatuses = make(map[int]*ConstantBase)

	// get all purchase order status from model purchase order
	var AllTransactionStatus = constant_models.GetTransactionStatuses()

	for constantKey, values := range AllTransactionStatus {
		PoCStatuses[constantKey] = &ConstantBase{
			Name:   values,
			Status: constantKey,
		}

		properties, err := GetTransactionProperties(constantKey)
		if err != nil {
			return nil, err
		}

		PoCStatuses[constantKey].Properties = properties
	}

	return PoCStatuses, nil
}

/**
Get Transaction status properties
- this function handling, create new object, get config email, etc
- it's entity editable or deletable
*/
func GetTransactionProperties(constantKey int) (properties StateProperties, err error) {
	deletableProperties, err := new_object.BaseCreateNewObject(constant.DeleteConfig).GetNewObject()
	if err != nil {
		return
	}

	// Set Properties
	switch constantKey {
	case constant.TransactionNew:
		properties = StateProperties{
			StateStatus:     constant.TransactionNew,
			Editable:        false,
			Deletable:       deletableProperties,
			CreateEmail:     nil,
			CreateNewObject: nil,
		}

	case constant.TransactionOnProgress:
		newObjectBase := new_object.BaseCreateNewObject(constant.Transactions)
		newObject, err := newObjectBase.GetNewObject()
		if err != nil {
			return StateProperties{}, err
		}

		properties = StateProperties{
			StateStatus: constant.TransactionOnProgress,
			Editable:    false,
			Deletable:   new_object.CreateNewObject{},
			CreateEmail: nil,
			CreateNewObject: []new_object.CreateNewObject{
				newObject,
			},
		}

	case constant.TransactionDone:
		properties = StateProperties{}

	case constant.TransactionComplete:
		properties = StateProperties{}

	default:
		err = fmt.Errorf("transaction status value [%d], not exists", constantKey)
	}

	if err != nil {
		return
	}

	emailBase := emails.NewEmailConfiguration(constant.Transactions, int8(constantKey))
	if err = emailBase.GetTransactionEmailConfiguration(); err != nil {
		return
	}

	properties.CreateEmail = emailBase.EmailData

	return
}

// ==================== Configuration EVENT, Transition Name, And State =====================

/**
Descriptions : this is for assign event, transitions and state
Note :
	- Create different transitions name, this for making event callback easier
*/
func (base *BaseSMTransaction) GetSMTransaction() (currentStateName string, stateEvents fsm.Events, stateCallbacks fsm.Callbacks, err error) {

	// check if state exists or not
	if _, err = base.getDetailStatus(base.CurrentStatus); err != nil {
		return
	}

	// get all po c status
	PoCStatus, _ := base.GetTransactionStatus()

	// assign current state name
	currentStateName = PoCStatus[base.CurrentStatus].Name // current state

	// assign event transitions
	stateEvents = fsm.Events{
		// normal transitions
		{
			Name: constant.TransactionConfirmed,                     // transitions name
			Src:  []string{PoCStatus[constant.TransactionNew].Name}, // source state name
			Dst:  PoCStatus[constant.TransactionOnProgress].Name,    // destination state name
		},
		{
			Name: constant.TransactionFinished,
			Src:  []string{PoCStatus[constant.TransactionOnProgress].Name},
			Dst:  PoCStatus[constant.TransactionDone].Name,
		},
		{
			Name: constant.TransactionCompleted,
			Src:  []string{PoCStatus[constant.TransactionDone].Name},
			Dst:  PoCStatus[constant.TransactionComplete].Name,
		},
	}

	// assign callback
	// this example you can change state to on progress,
	// but it will always can't change state to delivered
	stateCallbacks = fsm.Callbacks{
		"before_" + constant.TransactionConfirmed: func(e *fsm.Event) {
			state, err := base.TransactionConfirmed(e)
			if !state || err != nil {
				e.Cancel()
			}
		},
	}

	return
}

// ==================== Configuration EVENT CALLBACKS =====================

func (base *BaseSMTransaction) TransactionConfirmed(e *fsm.Event) (bool, error) {
	// code your check value or anything here

	if !base.Filtered {
		return true, nil
	}

	fmt.Println("Your current state " + base.Fsm.Current())
	return true, nil
}
