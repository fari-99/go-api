package models

type Companies struct {
	Base
	Name string `json:"name" gorm:"column:name"`
}

func (Companies) TableName() string {
	return "companies"
}
