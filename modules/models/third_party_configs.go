package models

import "time"

type ThirdPartyConfigs struct {
	ID        int64      `json:"id" gorm:"column:id"`
	CompanyID int64      `json:"company_id" gorm:"column:company_id"`
	Username  string     `json:"username" gorm:"username"`
	Password  string     `json:"password" gorm:"password"`
	SecretKey string     `json:"secret_key" gorm:"column:secret_key"`
	CreatedBy int64      `json:"created_by" gorm:"column:created_by"`
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"column:deleted_at"`

	Company Companies `json:"company" gorm:"foreignkey:CompanyID"`
}

func (ThirdPartyConfigs) TableName() string {
	return "third_party_configs"
}
