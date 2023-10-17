package models

type TwoAuths struct {
	Base
	UserID  IDType `gorm:"column:user_id" json:"user_id"`
	Account string `gorm:"column:account" json:"account"`
	Issuer  string `gorm:"column:issuer" json:"issuer"`
	Secret  string `gorm:"column:secret" json:"secret"`
	Status  int8   `gorm:"column:status" json:"status"`
}

func (Transactions) TwoAuths() string {
	return "two_fa"
}
