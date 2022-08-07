package models

type TwoAuthRecoveries struct {
	Base
	UserID string `gorm:"column:user_id" json:"user_id"`
	Code   string `gorm:"column:code" json:"code"`
	Status int8   `gorm:"column:status" json:"status"`
}

func (TwoAuthRecoveries) TwoAuths() string {
	return "two_fa_recoveries"
}
