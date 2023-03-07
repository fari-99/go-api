package models

type QuotationItems struct {
	Base
}

func (model QuotationItems) TableName() string {
	return "quotation_items"
}
