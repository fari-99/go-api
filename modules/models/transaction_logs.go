package models

type TransactionLogs struct {
	Base
}

func (model TransactionLogs) TableName() string {
	return "transaction_logs"
}
