package models

type Roles struct {
	Base
	RoleType int8   `gorm:"column:role_type" json:"role_type"`
	RoleName string `gorm:"column:role_name" json:"role_name"`
}

func (Roles) TableName() string {
	return "roles"
}
