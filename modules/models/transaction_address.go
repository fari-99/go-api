package models

// TODO : create user address

type TransactionAddress struct {
	Base
	TransactionID IDType  `json:"transaction_id"`
	UserType      int8    `json:"user_type"`    // customer / vendor
	AddressType   string  `json:"address_type"` // shipping / billing
	AddressID     IDType  `json:"address_id"`
	CountryID     IDType  `json:"country_id"`
	CountryName   string  `json:"country_name"`
	ProvinceID    IDType  `json:"province_id"`
	ProvinceName  string  `json:"province_name"`
	CityID        IDType  `json:"city_id"`
	CityName      string  `json:"city_name"`
	DistrictID    IDType  `json:"district_id"`
	DistrictName  string  `json:"district_name"`
	Address       string  `json:"address"`
	Phone         string  `json:"phone"`
	Fax           string  `json:"fax"`
	Mobile        string  `json:"mobile"`
	Postcode      string  `json:"postcode"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	Status        int8    `json:"status"`
	CreatedBy     IDType  `json:"created_by"`
}

func (model TransactionAddress) TableName() string {
	return "transaction_address"
}
