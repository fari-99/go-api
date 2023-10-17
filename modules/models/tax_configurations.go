package models

type TaxConfigurations struct {
	Base
	Name         string  `json:"name"`
	Label        string  `json:"label"`
	Code         string  `json:"code"`
	Descriptions string  `json:"descriptions"`
	Percentage   float64 `json:"percentage"`
	Status       int8    `json:"status"`
	CreatedBy    IDType  `json:"created_by"`
}

func (model TaxConfigurations) TableName() string {
	return "tax_configurations"
}
