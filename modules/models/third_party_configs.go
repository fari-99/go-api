package models

type ThirdPartyConfigs struct {
	Base
	CompanyID int64  `json:"company_id" gorm:"column:company_id"`
	Username  string `json:"username" gorm:"username"`
	Password  string `json:"password" gorm:"password"`
	SecretKey string `json:"secret_key" gorm:"column:secret_key"`
	CreatedBy int64  `json:"created_by" gorm:"column:created_by"`

	// relationship
	Company Companies `json:"company" gorm:"foreignkey:CompanyID"`
}

func (ThirdPartyConfigs) TableName() string {
	return "third_party_configs"
}
