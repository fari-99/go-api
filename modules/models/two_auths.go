package models

import "time"

type TwoAuths struct {
	ID     int64  `gorm:"column:id" json:"id"`
	UserID int64  `gorm:"column:user_id" json:"user_id"`
	Secret string `gorm:"column:secret" json:"secret"`
	Status int8   `gorm:"column:status" json:"status"`

	CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deleted_at" sql:"DEFAULT:NULL"`
}

func (Transactions) TwoAuths() string {
	return "transactions"
}
