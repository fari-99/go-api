package models

import "time"

// FileManagements is one processing job (summary, transcription, ...) for a storages file.
// Per-action data (params, results, errors) lives in FileManagementMeta.
type FileManagements struct {
	Base
	StorageID  IDType     `gorm:"column:storage_id" json:"storage_id"`
	Action     string     `gorm:"column:action" json:"action"`
	Status     int8       `gorm:"column:status" json:"status"`
	Attempt    int        `gorm:"column:attempt" json:"attempt"`
	StartedAt  *time.Time `gorm:"column:started_at" json:"started_at"`
	FinishedAt *time.Time `gorm:"column:finished_at" json:"finished_at"`
	CreatedBy  IDType     `gorm:"column:created_by" json:"created_by"`
}

func (FileManagements) TableName() string {
	return "file_managements"
}
