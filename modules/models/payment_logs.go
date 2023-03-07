package models

type PaymentLogs struct {
	Base
}

func (model PaymentLogs) TableName() string {
	return "payment_logs"
}
