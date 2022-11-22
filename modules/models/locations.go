package models

import (
	"gorm.io/gorm"

	"go-api/constant"
)

type Locations struct {
	Base
	ParentID     string `gorm:"column:parent_id" json:"parent_id"`
	Code         string `gorm:"column:code" json:"code"`
	Name         string `gorm:"column:name" json:"name"`
	CompleteName string `gorm:"column:complete_name" json:"complete_name"`
	LevelID      string `gorm:"column:level_id" json:"level_id"`
	Status       int8   `gorm:"column:status" json:"status"`

	// Relations
	Levels LocationLevels `gorm:"-" json:"levels"`
}

func (Locations) TableName() string {
	return "locations"
}

func (storage *Locations) BeforeDelete(tx *gorm.DB) (err error) {
	tx.Model(storage).Update("status", constant.StatusDeleted)
	return
}

// ---------------------------------

type LocationLevels struct {
	Base
	NeedParentID bool   `gorm:"column:need_parent_id" json:"need_parent_id"`
	Name         string `gorm:"column:name" json:"name"`
	Status       int8   `gorm:"column:status" json:"status"`
}

func (LocationLevels) TableName() string {
	return "location_levels"
}

func (storage *LocationLevels) BeforeDelete(tx *gorm.DB) (err error) {
	tx.Model(storage).Update("status", constant.StatusDeleted)
	return
}
