package models

import (
	"go-api/constant"
	"time"

	"github.com/jinzhu/gorm"
)

type Storages struct {
	Id               int32      `gorm:"primary_key" json:"id"`
	Type             string     `json:"type" binding:"required"`
	Path             string     `json:"path" binding:"required"`
	Filename         string     `json:"filename" binding:"required"`
	Mime             string     `json:"mime"`
	OriginalFilename string     `json:"original_filename"`
	Status           int8       `json:"status"`
	CreatedBy        int32      `json:"created_by"`
	CreatedAt        time.Time  `json:"created_at"`
	DeletedAt        *time.Time `sql:"DEFAULT:NULL" json:"deleted_at"`
}

func (storage *Storages) BeforeDelete(scope *gorm.Scope) (err error) {
	scope.DB().Model(storage).Update("status", constant.StatusDeleted)
	return
}
