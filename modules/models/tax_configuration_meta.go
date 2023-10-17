package models

type TaxConfigurationMeta struct {
	Base
	TaxConfigurationID IDType `json:"tax_configuration_id"`
	Group              string `json:"group"` // ex: options
	Key                string `json:"key"`   // ex: is_editable, is_shown, is_local_tax, is_government_tax
	Value              string `json:"value"` // ex: true, false, [1..9]+, [a-zA-Z]+
	CreatedBy          IDType `json:"created_by"`
}

func (model TaxConfigurationMeta) TableName() string {
	return "tax_configuration_meta"
}
