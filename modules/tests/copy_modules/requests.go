package copy_modules

type RequestListFilter struct {
	Page    int  `json:"page"`
	Limit   int  `json:"limit"`
	OrderBy string `json:"order_by"`
}
