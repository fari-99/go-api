package models

type Quotations struct {
	Base
}

func (model Quotations) TableName() string {
	return "quotations"
}
