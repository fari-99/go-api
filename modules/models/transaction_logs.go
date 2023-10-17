package models

type TransactionLogs struct {
	Base
	TransactionID IDType `json:"transaction_id"`
	ModelName     string `json:"entity"`
	ModelID       IDType `json:"entity_id"`
	Actor         string `json:"actor"`
	Title         string `json:"title"`
	Message       string `json:"message"`
}

func (model TransactionLogs) TableName() string {
	return "transaction_logs"
}
