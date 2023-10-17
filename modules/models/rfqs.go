package models

type Rfqs struct {
	Base
	TransactionID IDType `json:"transaction_id"`
	CompanyID     IDType `json:"company_id"`

	RfqNumber string `json:"rfq_number"`
	Notes     string `json:"notes"`
	Reference string `json:"reference"`

	Status       int8   `json:"status"`
	StatusReason string `json:"status_reason"`

	SubTotal float64 `json:"sub_total"`
	TaxTotal float64 `json:"tax_total"`
	Total    float64 `json:"total"`
}

func (model Rfqs) TableName() string {
	return "rfqs"
}
