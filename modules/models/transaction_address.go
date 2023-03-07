package models

type TransactionAddress struct {
	Base
}

func (model TransactionAddress) TableName() string {
	return "transaction_address"
}
