package models

import "time"

type Users struct {
	Base
	Username     string `gorm:"column:username" json:"username"`
	Password     string `gorm:"column:password" json:"-"`
	Email        string `gorm:"column:email" json:"email"`
	Status       int8   `gorm:"column:status" json:"status"`
	MobilePhone  string `gorm:"column:mobile_phone" json:"mobile_phone"`
	TwoFaEnabled bool   `gorm:"column:two_fa_enabled" json:"two_fa_enabled"`

	Roles       string        `json:"roles" gorm:"-"`
	UserSocials []UserSocials `json:"user_socials" gorm:"foreignkey:UserID"`

	TwoFaModels *TwoAuthsModels `gorm:"-" json:"two_fa_models"`
}

func (Users) TableName() string {
	return "users"
}

type TwoAuthsModels struct {
	TOTP         bool `gorm:"column:-" json:"totp"`
	RecoveryCode bool `gorm:"column:-" json:"recovery_code"`
	Email        bool `gorm:"column:-" json:"email"`
}

type UserProfile struct {
	Username     string `json:"username"`
	Email        string `json:"email"`
	Status       int8   `json:"status"`
	MobilePhone  string `json:"mobile_phone"`
	TwoFaEnabled bool   `json:"two_fa_enabled"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`

	TwoFaModels *TwoAuthsModels `gorm:"-" json:"two_fa_models"`
}
