package constant_models

import "go-api/constant"

func GetTransactionStatuses() map[int]string {
	transactionStatus := make(map[int]string)

	transactionStatus[constant.TransactionNew] = "New"
	transactionStatus[constant.TransactionOnProgress] = "On Progress"
	transactionStatus[constant.TransactionDone] = "Need Payment"
	transactionStatus[constant.TransactionComplete] = "Completed"

	return transactionStatus
}
