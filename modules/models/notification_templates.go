package models

import "time"

type NotificationTemplates struct {
	ID               int64  `gorm:"primary_key:id" json:"id"`
	NotificationType int8   `gorm:"column:notification_type" json:"notification_type"`
	Action           string `gorm:"column:action" json:"action"`
	Subject          string `gorm:"column:subject" json:"subject"`
	Body             string `gorm:"column:body" json:"body"`
	Status           int8   `gorm:"column:status" json:"status"`

	CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deleted_at" sql:"DEFAULT:NULL"`
}

func (NotificationTemplates) TableName() string {
	return "notification_templates"
}
