package models

type TransactionActionLogs struct {
	Base
	ModelName  string `json:"model_name"` // table_name
	ModelID    IDType `json:"model_id"`
	ActionName string `json:"action"` // ex: rfq-created, etc
	ActionBy   IDType `json:"action_by"`
}

func (model TransactionActionLogs) TableName() string {
	return "transaction_action_logs"
}
