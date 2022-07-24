package models

import "time"

type TwoAuthRecoveries struct {
	ID     uint64 `gorm:"column:id" json:"id"`
	UserID uint64 `gorm:"column:user_id" json:"user_id"`
	Code   string `gorm:"column:code" json:"code"`
	Status int8   `gorm:"column:status" json:"status"`

	CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deleted_at" sql:"DEFAULT:NULL"`
}

func (TwoAuthRecoveries) TwoAuths() string {
	return "two_fa_recoveries"
}
