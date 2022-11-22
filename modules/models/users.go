package models

import "time"

type Users struct {
	Base
	Username     string `gorm:"column:username" json:"username"`
	Password     string `gorm:"column:password" json:"password"`
	Email        string `gorm:"column:email" json:"email"`
	Status       int8   `gorm:"column:status" json:"status"`
	MobilePhone  string `gorm:"column:mobile_phone" json:"mobile_phone"`
	TwoFaEnabled bool   `gorm:"column:two_fa_enabled" json:"two_fa_enabled"`

	Roles       string        `json:"roles" gorm:"-"`
	UserSocials []UserSocials `json:"user_socials" gorm:"foreignkey:UserID"`
}

func (Users) TableName() string {
	return "users"
}

type UserProfile struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Status   int8   `json:"status"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}
