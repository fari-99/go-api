package models

type TransactionRelations struct {
	Base
}

func (model TransactionRelations) TableName() string {
	return "transaction_relations"
}
