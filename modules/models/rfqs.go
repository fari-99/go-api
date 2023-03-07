package models

type Rfqs struct {
	Base
}

func (model Rfqs) TableName() string {
	return "rfqs"
}
