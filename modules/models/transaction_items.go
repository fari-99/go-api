package models

type TransactionItems struct {
	Base
	TransactionID string `gorm:"column:transaction_id" json:"transaction_id"`
	Status        uint8  `gorm:"column:status" json:"status"`
	CreatedBy     uint64 `gorm:"column:created_by" json:"created_by"`
}

func (Transactions) TransactionItems() string {
	return "transaction_items"
}
