package models

import (
	"go-api/pkg/go2rtc_helper"
)

type SecurityCameras struct {
	Base
	UserID       uint64 `gorm:"column:user_id" json:"user_id"`
	Name         string `gorm:"column:name" json:"name"`
	Uuid         string `gorm:"column:uuid" json:"uuid"`
	Url          string `gorm:"column:url" json:"url"`
	Username     string `gorm:"column:username" json:"username"`
	Password     string `gorm:"column:password" json:"password"`
	Descriptions string `gorm:"column:descriptions" json:"descriptions"`

	// go2rtc data
	Stream go2rtc_helper.Stream `gorm:"-" json:"stream"`
}

func (SecurityCameras) TableName() string {
	return "security_cameras"
}
