package file_managements

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	"go-api/modules/models"
)

type CreateFileManagementRequest struct {
	StorageID uint64         `json:"storage_id"`
	Action    string         `json:"action"`
	Params    map[string]any `json:"params"`
}

func (m CreateFileManagementRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.StorageID, validation.Required),
		validation.Field(&m.Action, validation.Required, validation.Length(1, 50)),
	)
}

// ListFilter of GET /file-managements, CreatedBy is always the current user
type ListFilter struct {
	Page      int
	Limit     int
	Status    int8
	Action    string
	StorageID uint64
	CreatedBy uint64
}

type ActionInfo struct {
	Action             string      `json:"action"`
	AcceptedMimes      []string    `json:"accepted_mimes"`
	AcceptedExtensions []string    `json:"accepted_extensions"`
	MaxFileBytes       int64       `json:"max_file_bytes"`
	Params             []ParamSpec `json:"params"`
}

type StorageSummary struct {
	ID               models.IDType `json:"id"`
	OriginalFilename string        `json:"original_filename"`
	Mime             string        `json:"mime"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type FileManagementItem struct {
	ID              models.IDType `json:"id"`
	StorageID       models.IDType `json:"storage_id"`
	Action          string        `json:"action"`
	Status          int8          `json:"status"`
	StatusLabel     string        `json:"status_label"`
	AvailableEvents []string      `json:"available_events"`
	Attempt         int           `json:"attempt"`
	StartedAt       *time.Time    `json:"started_at"`
	FinishedAt      *time.Time    `json:"finished_at"`
	CreatedBy       models.IDType `json:"created_by"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

type FileManagementDetail struct {
	FileManagementItem
	Storage *StorageSummary                         `json:"storage"`
	Meta    map[string]map[string]map[string]string `json:"meta"`
	Error   *ErrorInfo                              `json:"error"`
}
