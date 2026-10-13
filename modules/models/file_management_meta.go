package models

type FileManagementMeta struct {
	Base
	FileManagementID IDType `gorm:"column:file_management_id" json:"file_management_id"`
	Group            string `gorm:"column:group" json:"group"` // ex: input, output, request, error
	Type             string `gorm:"column:type" json:"type"`   // sub category of group, ex: params, summary, attempt_1
	Key              string `gorm:"column:key" json:"key"`
	Value            string `gorm:"column:value" json:"value"` // always a string, non-string data is JSON encoded
	CreatedBy        IDType `gorm:"column:created_by" json:"created_by"`
}

func (FileManagementMeta) TableName() string {
	return "file_management_meta"
}
