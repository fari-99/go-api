package models

type UserRoles struct {
	Base
	UserID uint64 `gorm:"column:user_id" json:"user_id"`
	RoleID uint64 `gorm:"column:role_id" json:"role_id"`
}

func (UserRoles) TableName() string {
	return "user_roles"
}
