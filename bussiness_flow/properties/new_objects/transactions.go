package new_object

import "go-api/constant"

func (base *BaseNewObject) NewTransactionData() (NewObject CreateNewObject) {
	NewObject = CreateNewObject{
		ObjectName:  constant.Transactions,
		StatusValue: constant.TransactionNew,
	}

	return
}

func (base *BaseNewObject) DeleteTransactionData() (NewObject CreateNewObject) {
	NewObject = CreateNewObject{
		ObjectName:  constant.Transactions,
		StatusValue: constant.TransactionDelete,
	}

	return
}
