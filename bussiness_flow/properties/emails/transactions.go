package emails

import (
	"fmt"
	"go-api/constant"
)

func (base *EmailConfigurationBase) GetTransactionEmailConfiguration() (err error) {
	var emailData []EmailData

	switch base.StatusData {
	case constant.TransactionNew:
		emailData = []EmailData{
			{
				SendTo:     SendToAllCustomer,
				ActionName: TransactionNewCustomer,
				UserType:   constant.UserTypeCustomer,
			},
		}

	case constant.TransactionOnProgress:
		emailData = []EmailData{}

	case constant.TransactionDone:
		emailData = []EmailData{}

	case constant.TransactionComplete:
		emailData = []EmailData{}

	default:
		err = fmt.Errorf("transaction status [%d] don't have email configuration", base.StatusData)
	}

	if err != nil {
		return
	}

	base.EmailData = emailData

	return
}
