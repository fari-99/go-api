package emails

import (
	"errors"
	"go-api/constant"
)

type EmailConfigurationBase struct {
	DataType   string `json:"data_type"`
	StatusData int    `json:"status_data"`

	// accumulate error
	Errors []error `json:"errors"`

	// Data Configuration
	EmailData []EmailData `json:"email_data"`
}

type EmailData struct {
	SendTo     string `json:"send_to"`
	ActionName string `json:"action_name"`
	UserType   int    `json:"user_type"`
}

func NewEmailConfiguration(dataType string, statusData int8) *EmailConfigurationBase {

	emailBase := &EmailConfigurationBase{
		DataType:   dataType,
		StatusData: int(statusData),
	}

	return emailBase
}

func (base *EmailConfigurationBase) GetEmailConfiguration() *EmailConfigurationBase {
	var err error

	switch base.DataType {
	case constant.Transactions:
		err = base.GetTransactionEmailConfiguration()
	default:
		err = errors.New("data type not supported")
	}

	if err != nil {
		base.Errors = append(base.Errors, err)
	}

	return base
}
