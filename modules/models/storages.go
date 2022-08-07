package models

import (
	"go-api/constant"

	"gorm.io/gorm"
)

type Storages struct {
	Base
	Type             string `gorm:"column:type" json:"type"`
	Path             string `gorm:"column:path" json:"path"`
	Filename         string `gorm:"column:filename" json:"filename"`
	Mime             string `gorm:"column:mime" json:"mime"`
	OriginalFilename string `gorm:"column:original_filename" json:"original_filename"`
	Status           int8   `gorm:"column:status" json:"status"`
	CreatedBy        string `gorm:"column:created_by" json:"created_by"`
}

func (Storages) TableName() string {
	return "storages"
}

func (storage *Storages) BeforeDelete(tx *gorm.DB) (err error) {
	tx.Model(storage).Update("status", constant.StatusDeleted)
	return
}
