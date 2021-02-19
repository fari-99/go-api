package models

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"time"
)

type SocialMedia struct {
	ID        int64      `gorm:"primary_key:id" json:"id"`
	UserID    int64      `gorm:"column:user_id" json:"user_id"`
	Uuid      string     `gorm:"column:uuid" json:"uuid"`
	Name      string     `gorm:"column:name" json:"name"`
	TokenID   string     `gorm:"column:token_id" json:"token_id"`
	Status    int8       `gorm:"column:status" json:"status"`
	CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deleted_at" sql:"DEFAULT:NULL"`
}

func (SocialMedia) TableName() string {
	return "social_media"
}

func (model *SocialMedia) BeforeCreate(tx *gorm.DB) (err error) {
	model.Uuid = uuid.New().String()
	return nil
}
