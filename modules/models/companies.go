package models

import "time"

type Companies struct {
	ID        int64      `json:"id" gorm:"column:id"`
	Name      string     `json:"name" gorm:"column:name"`
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"column:deleted_at"`
}

func (Companies) TableName() string {
	return "companies"
}
