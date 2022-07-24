package models

import "time"

type Roles struct {
	ID        uint64     `gorm:"primary_key:id" json:"id"`
	RoleType  int8       `gorm:"column:role_type" json:"role_type"`
	RoleName  string     `gorm:"column:role_name" json:"role_name"`
	CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deleted_at"`
}

func (Roles) TableName() string {
	return "roles"
}
