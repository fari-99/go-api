package models

type UserRoles struct {
	Base
	UserID IDType `gorm:"column:user_id" json:"user_id"`
	RoleID IDType `gorm:"column:role_id" json:"role_id"`
}

func (UserRoles) TableName() string {
	return "user_roles"
}
