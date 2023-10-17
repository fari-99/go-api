package models

type TransactionRelations struct {
	Base
	TransactionID string `json:"transaction_id"`
	ModelName     string `json:"entity_name"`
	ModelID       string `json:"entity_id"`
	CreatedBy     string `json:"created_by"`

	ModelParentName string `json:"entity_parent_name,omitempty"`
	ModelParentID   string `json:"entity_parent_id,omitempty"`
	CreatedParentBy string `json:"created_parent_by,omitempty"`
}

func (model TransactionRelations) TableName() string {
	return "transaction_relations"
}
