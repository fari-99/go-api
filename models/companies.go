package models

import "time"

type Companies struct {
	ID          int64  `gorm:"primary_key:id" json:"id"`
	Name        string `gorm:"column:name" json:"name"`
	Code        string `gorm:"column:code" json:"code"`
	ParentID    int64  `gorm:"column:parent_id" json:"parent_id"`
	CompanyType int8   `gorm:"column:company_type" json:"company_type"`
	Status      int8   `gorm:"column:status" json:"status"`

	CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deleted_at" sql:"DEFAULT:NULL"`
}
