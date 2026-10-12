package models

import (
	"time"

	"github.com/spf13/cast"
	"gorm.io/gorm"
)

type IDType uint64

func (id IDType) String() string {
	return cast.ToString(uint64(id))
}

func (id IDType) Uint64() uint64 {
	return uint64(id)
}

type Base struct {
	ID        IDType          `gorm:"column:id" json:"id" sql:"type:uuid;primary_key;default:uuid_generate_v4()" `
	CreatedAt time.Time       `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time       `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *gorm.DeletedAt `gorm:"column:deleted_at" json:"deleted_at" swaggertype:"string" format:"date-time" sql:"DEFAULT:NULL"`
}

func (base *Base) BeforeCreate(tx *gorm.DB) error {
	return nil
}

func (base *Base) AfterCreate(tx *gorm.DB) error {
	return nil
}

func (base *Base) AfterUpdate(tx *gorm.DB) error {
	return nil
}

func (base *Base) AfterDelete(tx *gorm.DB) error {
	return nil
}
