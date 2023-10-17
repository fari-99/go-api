package models

type TransactionDocuments struct {
	Base
	DocumentName string `json:"document_name"`
	DocumentType string `json:"document_type"`
	StorageID    IDType `json:"storage_id"`
	Filename     string `json:"filename"`
	CreatedBy    IDType `json:"created_by"`
}

func (model TransactionDocuments) TableName() string {
	return "transaction_documents"
}
