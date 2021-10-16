package models

import (
	"go-api/constant"
	"time"

	"github.com/jinzhu/gorm"
)

type Storages struct {
	Id               int64      `gorm:"primary_key:id" json:"id"`
	Type             string     `gorm:"column:type" json:"type"`
	Path             string     `gorm:"column:path" json:"path"`
	Filename         string     `gorm:"column:filename" json:"filename"`
	Mime             string     `gorm:"column:mime" json:"mime"`
	OriginalFilename string     `gorm:"column:original_filename" json:"original_filename"`
	Status           int8       `gorm:"column:status" json:"status"`
	CreatedBy        int32      `gorm:"column:created_by" json:"created_by"`
	CreatedAt        time.Time  `gorm:"column:created_at" json:"created_at"`
	DeletedAt        *time.Time `gorm:"column:deleted_at" json:"deleted_at"`
}

func (Storages) TableName() string {
	return "storages"
}

func (storage *Storages) BeforeDelete(scope *gorm.Scope) (err error) {
	scope.DB().Model(storage).Update("status", constant.StatusDeleted)
	return
}
