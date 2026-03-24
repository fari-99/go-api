package models

type TwoAuths struct {
	Base
	UserID   IDType `gorm:"column:user_id" json:"user_id"`
	Type     string `gorm:"column:type" json:"type"`           // totp, otp
	SendType string `gorm:"column:send_type" json:"send_type"` // authenticator, email, whatsapp, sms, etc
	Account  string `gorm:"column:account" json:"account"`
	Issuer   string `gorm:"column:issuer" json:"issuer"`
	Secret   string `gorm:"column:secret" json:"secret"`
	Status   int8   `gorm:"column:status" json:"status"`
}

func (TwoAuths) TableName() string {
	return "two_fa"
}
