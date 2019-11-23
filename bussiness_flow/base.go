package bussiness_flow

import (
	"errors"
	"fmt"
	"go-api/bussiness_flow/properties/emails"
	"go-api/bussiness_flow/properties/new_objects"
	"go-api/constant"
	"reflect"

	"github.com/looplab/fsm"
)

type BaseSMTransaction struct {
	ConfigName     string                 `json:"configName"`
	Filtered       bool                   `json:"filtered"`
	TransitionName string                 `json:"transitionName,omitempty"`
	Data           map[string]interface{} `json:"data"`
	CurrentStatus  int                    `json:"current_status"`

	Fsm       *fsm.FSM
	ErrorData []string `json:"error_data"`
}

// Transitions Finite State Machine Struct
type TransactionTransitions struct {
	TransitionName string
	NewStateName   string
	NewStateValue  int
}

type ConstantBase struct {
	Name       string
	Status     int
	Properties StateProperties
}

type StateProperties struct {
	StateStatus int `json:"state_status,omitempty"`

	//UpdateItem []updateItems.UpdateItems `json:"update_item,omitempty"`

	Editable  bool                       `json:"editable,omitempty"`
	Deletable new_object.CreateNewObject `json:"deletable,omitempty"`

	CreateEmail     []emails.EmailData           `json:"create_email,omitempty"`
	CreateNewObject []new_object.CreateNewObject `json:"create_new_object,omitempty"`
}

type Properties map[string]interface{}

func NewBaseStateMachine(configName string, isFiltered bool, transitionName string, data map[string]interface{}) (*BaseSMTransaction, error) {
	base := &BaseSMTransaction{
		ConfigName:     configName,
		Filtered:       isFiltered,
		TransitionName: transitionName,
		Data:           data,
		CurrentStatus:  int(reflect.ValueOf(data["status"]).Float()),
	}

	err := base.GetEventState()
	if err != nil {
		return &BaseSMTransaction{}, err
	}

	return base, nil
}

func (base *BaseSMTransaction) GetEventState() (err error) {
	var currentStateName string
	var stateEvents fsm.Events
	var stateCallbacks fsm.Callbacks

	// please add your configName parameter here
	switch base.ConfigName {
	case constant.Transactions:
		currentStateName, stateEvents, stateCallbacks, err = base.GetSMTransaction()
	default:
		err = fmt.Errorf("can't get available transition, no name for that configuration name [%s]", base.TransitionName)
	}

	if err != nil {
		return
	}

	base.Fsm = fsm.NewFSM(
		currentStateName,
		stateEvents,
		stateCallbacks,
	)

	return
}

func (base *BaseSMTransaction) ChangeStateMachine() (success bool, err error) {
	success = false // assume all state can't change state

	currentState := base.Fsm.Current()

	transitionName := base.TransitionName

	err = base.Fsm.Event(transitionName)

	if currentState == base.Fsm.Current() {
		return
	}

	success = true
	return
}

func (base *BaseSMTransaction) GetAvailableTransitions() (transitions map[string]TransactionTransitions, err error) {
	var constantBase = make(map[int]*ConstantBase)
	transitions = make(map[string]TransactionTransitions)

	// get all available transitions
	availableTransitions := base.Fsm.AvailableTransitions()

	// please add your configName parameter here
	switch base.ConfigName {
	case constant.Transactions:
		constantBase, err = base.GetTransactionStatus()
	default:
		err = fmt.Errorf("can't get available transition, no name for that configuration name [%s]", base.TransitionName)
	}

	if err != nil {
		return
	}

	for _, transitionName := range availableTransitions {
		var currentState string

		// set current state before filtered
		currentState = constantBase[base.CurrentStatus].Name

		// set state using current state
		base.Fsm.SetState(currentState)

		err = base.Fsm.Event(transitionName)
		if err != nil {
			return
		}

		nextState := base.Fsm.Current()

		if currentState == nextState {
			continue
		}

		currentState = nextState

		constraitValue, err := base.GetStatusByName(base.Fsm.Current())
		if err != nil {
			continue
		}

		transitions[transitionName] = TransactionTransitions{
			TransitionName: transitionName,
			NewStateName:   currentState,
			NewStateValue:  constraitValue,
		}
	}

	return transitions, nil
}

func (base *BaseSMTransaction) GetStateProperties() (properties StateProperties, err error) {

	allStatus, err := base.getAllStatus()
	if err != nil {
		return
	}

	status := base.CurrentStatus

	if _, ok := allStatus[status]; !ok {
		err = fmt.Errorf("can't get state properties, status [%d] not exists", status)
		return
	}

	properties = allStatus[status].Properties
	return
}

func (base *BaseSMTransaction) getAllStatus() (allStatus map[int]*ConstantBase, err error) {
	allStatus = make(map[int]*ConstantBase)

	// please add your configName parameter here
	switch base.ConfigName {
	case constant.Transactions:
		allStatus, err = base.GetTransactionStatus()
	default:
		err = fmt.Errorf("can't get available transition, no name for that configuration name [%s]", base.TransitionName)
		return
	}

	return
}

func (base *BaseSMTransaction) getDetailStatus(status int) (statusDetail *ConstantBase, err error) {
	allStatus, err := base.getAllStatus()
	if err != nil {
		return
	}

	if _, ok := allStatus[status]; !ok {
		err = fmt.Errorf("can't get detail status, status [%d] not exists", status)
		return &ConstantBase{}, err
	}

	statusDetail = allStatus[status]

	return
}

func (base *BaseSMTransaction) GetStatusByName(constraitName string) (status int, err error) {
	allStatus, err := base.getAllStatus()
	if err != nil {
		return
	}

	status = 0
	for key, value := range allStatus {
		if value.Name == constraitName {
			status = key
			return
		}
	}

	if status == 0 {
		err = errors.New("constrait name is invalid")
		return
	}

	return
}
