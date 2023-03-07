package models

type OrderItems struct {
	Base
}

func (model OrderItems) TableName() string {
	return "order_items"
}
