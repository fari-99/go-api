package models

type RfqItems struct {
	Base
}

func (model RfqItems) TableName() string {
	return "rfq_items"
}
