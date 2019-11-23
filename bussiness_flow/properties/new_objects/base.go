package new_object

import (
	"fmt"
	"go-api/constant"
)

type BaseNewObject struct {
	NewObjectName string
}

// this will handle create quotation, and create purchase order
type CreateNewObject struct {
	ObjectName      string `json:"object_name,omitempty"`
	StatusValue     int    `json:"status_value,omitempty"`
	ItemStatusValue int    `json:"item_status_value,omitempty"`
	TransitionName  string `json:"transition_name,omitempty"`
}

func BaseCreateNewObject(NewObjectName string) (baseObjectName *BaseNewObject) {

	baseObjectName = &BaseNewObject{
		NewObjectName: NewObjectName,
	}

	return
}

func (base *BaseNewObject) GetNewObject() (NewObject CreateNewObject, err error) {

	switch base.NewObjectName {
	case constant.Transactions:
		NewObject = base.NewTransactionData()
	case constant.DeleteConfig:
		NewObject = base.DeleteTransactionData()
	default:
		err = fmt.Errorf("object name [%s] doesn't exists", base.NewObjectName)
	}

	return
}
