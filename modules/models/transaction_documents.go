package models

type TransactionDocuments struct {
	Base
}

func (model TransactionDocuments) TableName() string {
	return "transaction_documents"
}
