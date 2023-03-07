package models

type CartItems struct {
	Base
}

func (model CartItems) TableName() string {
	return "cart_items"
}
