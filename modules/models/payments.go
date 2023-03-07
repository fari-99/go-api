package models

type Payments struct {
	Base
}

func (model Payments) TableName() string {
	return "payments"
}
