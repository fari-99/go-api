package models

type Articles struct {
}

func (Articles) TableName() string {
	return "articles"
}
