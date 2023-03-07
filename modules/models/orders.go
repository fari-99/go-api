package models

type Orders struct {
	Base
}

func (model Orders) TableName() string {
	return "orders"
}
