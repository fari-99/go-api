package models

import "time"

type CalendarManagements struct {
	Base
	CountryID    IDType    `gorm:"column:country_id" json:"country_id"`
	Dates        time.Time `gorm:"column:dates" json:"dates"`
	Descriptions string    `gorm:"column:descriptions" json:"descriptions"`
	CreatedBy    IDType    `gorm:"column:created_by" json:"created_by"`
	Status       int8      `gorm:"column:status" json:"status"`
}

func (CalendarManagements) TableName() string {
	return "calendar_managements"
}
