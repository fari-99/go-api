package faspay

import "fmt"

const ResponseCodeSuccess = "00"
const ResponseCodeInvalidMerchant = "03"
const ResponseCodeInvalidAmount = "13"
const ResponseCodeInvalidOrder = "14"
const ResponseCodeOrderCancelled = "17"
const ResponseCodeInvalidCustomer = "18"
const ResponseCodeSubsExpired = "21"
const ResponseCodeFormatError = "30"
const ResponseCodeRequestedFunctionNotSupported = "40"
const ResponseCodeOrderExpired = "54"
const ResponseCodeIncorrectUser = "55"
const ResponseCodeSecurityViolation = "56"
const ResponseCodeNotActive = "63"
const ResponseCodeInternalError = "66"
const ResponseCodePaymentWasReversal = "80"
const ResponseCodeAlreadyPaid = "81"
const ResponseCodeUnregisteredEntity = "82"
const ResponseCodeParameterMandatory = "83"
const ResponseCodeUnregisteredParameter = "84"
const ResponseCodeInsufficientParameter = "85"
const ResponseCodeSystemMalfunction = "96"

func AllLabelResponseCode() map[string]string {
	mapResponseCode := map[string]string{
		ResponseCodeSuccess:                       "Success",
		ResponseCodeInvalidMerchant:               "Invalid Merchant",
		ResponseCodeInvalidAmount:                 "Invalid Amount",
		ResponseCodeInvalidOrder:                  "Invalid Order",
		ResponseCodeOrderCancelled:                "Order Cancelled by Merchant/Customer",
		ResponseCodeInvalidCustomer:               "Invalid Customer or MSISDN is not found",
		ResponseCodeSubsExpired:                   "Subscription is Expired",
		ResponseCodeFormatError:                   "Format Error",
		ResponseCodeRequestedFunctionNotSupported: "Requested Function not Supported",
		ResponseCodeOrderExpired:                  "Order is Expired",
		ResponseCodeIncorrectUser:                 "Incorrect User/Password",
		ResponseCodeSecurityViolation:             "Security Violation (from unknown IP-Address)",
		ResponseCodeNotActive:                     "Not Active / Suspended",
		ResponseCodeInternalError:                 "Internal Error",
		ResponseCodePaymentWasReversal:            "Payment Was Reversal",
		ResponseCodeAlreadyPaid:                   "Already Been Paid",
		ResponseCodeUnregisteredEntity:            "Unregistered Entity",
		ResponseCodeParameterMandatory:            "Parameter is mandatory",
		ResponseCodeUnregisteredParameter:         "Unregistered Parameters",
		ResponseCodeInsufficientParameter:         "Insufficient Paramaters",
		ResponseCodeSystemMalfunction:             "System Malfunction",
	}

	return mapResponseCode
}

func GetLabelResponseCode(responseCode string) (string, error) {
	mapResponseCode := AllLabelResponseCode()
	if _, ok := mapResponseCode[responseCode]; ok {
		return mapResponseCode[responseCode], nil
	}

	// https://docs.faspay.co.id/merchant-integration/api-reference-1/debit-transaction/reference/status-response-code
	return "", fmt.Errorf("response code not set, please check faspay documentation for any response code update")
}
