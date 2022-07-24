package models

import "time"

type TransactionItems struct {
	ID            uint64 `gorm:"column:id" json:"id"`
	TransactionID string `gorm:"column:transaction_id" json:"transaction_id"`
	Status        uint8  `gorm:"column:status" json:"status"`
	CreatedBy     uint64 `gorm:"column:created_by" json:"created_by"`

	CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deleted_at" sql:"DEFAULT:NULL"`
}

func (Transactions) TransactionItems() string {
	return "transaction_items"
}
