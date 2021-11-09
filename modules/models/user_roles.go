package models

import "time"

type UserRoles struct {
	ID        uint64     `gorm:"primary_key:id" json:"id"`
	UserID    uint64     `gorm:"column:user_id" json:"user_id"`
	RoleID    uint64     `gorm:"column:role_id" json:"role_id"`
	CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deleted_at"`
}

func (model UserRoles) TableName() string {
	return "user_roles"
}
