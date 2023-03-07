package models

type Carts struct {
	Base
}

func (model Carts) TableName() string {
	return "carts"
}
