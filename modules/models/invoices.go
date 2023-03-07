package models

type Invoices struct {
	Base
}

func (model Invoices) TableName() string {
	return "invoices"
}
