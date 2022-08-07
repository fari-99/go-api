package models

type Articles struct {
	Base
}

func (Articles) TableName() string {
	return "articles"
}
