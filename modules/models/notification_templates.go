package models

type NotificationTemplates struct {
	Base
	NotificationType int8   `gorm:"column:notification_type" json:"notification_type"`
	Action           string `gorm:"column:action" json:"action"`
	Subject          string `gorm:"column:subject" json:"subject"`
	Body             string `gorm:"column:body" json:"body"`
	Status           int8   `gorm:"column:status" json:"status"`
}

func (NotificationTemplates) TableName() string {
	return "notification_templates"
}
