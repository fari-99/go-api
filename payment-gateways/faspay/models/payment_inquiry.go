package models

// M : Mandatory
// O : Optional
// C : Conditional

/*
PaymentInquiryRequest
|Parameter  |Data Type       |M/O/C|Description                |
|-----------|----------------|-----|---------------------------|
|Request    |Alfanumeric (50)|O    |Request Description        |
|Merchant_id|Numeric (5)     |M    |Merchant Code              |
|Merchant   |Alfanumeric (32)|O    |Merchant Name              |
|Signature  |Alfanumeric     |M    |sha1(md5(user_id+password))|

Example PaymentInquiryRequest:
{
    "request":"Request List of Payment Gateway",
    "merchant_id":"98765",
    "merchant":"FASPAY STORE",
    "signature":"e73f632e99308e424cac281b5d28c4cfc150e85b"
}
*/
type PaymentInquiryRequest struct {
	Request    string `json:"request"`
	MerchantID string `json:"merchant_id"`
	Merchant   string `json:"merchant"`
	Signature  string `json:"signature"`
}

/*
PaymentInquiryResponse
| Parameter     | Data Type        | M/O/C | Description          |
|---------------|------------------|-------|----------------------|
| Response      | Alfanumeric (50) | O     | Response Description |
| Merchant_id   | Numeric (5)      | M     | Merchant Code        |
| Merchant      | Alfanumeric (32) | O     | Merchant Name        |
| PG code       | Numeric (3)      | M     | Payment Channel Code |
| PG name       | Alfanumeric (32) | M     | Payment Channel Name |
| Response_Code | Alfanumeric (2)  | M     | Response Code        |
| Response_Desc | Alfanumeric (32) | M     | Response Description |

Example PaymentInquiryResponse:
{
    "response": "List Payment Channel",
    "merchant_id": "99998",
    "merchant": "Rizki Store",
    "payment_channel": [
        {
            "pg_code": "707",
            "pg_name": "ALFAGROUP"
        }
    ],
    "response_code": "00",
    "response_desc": "Success"
}
*/
type PaymentInquiryResponse struct {
	Response       string `json:"response"`
	MerchantID     string `json:"merchant_id"`
	Merchant       string `json:"merchant"`
	PaymentChannel []struct {
		PgCode string `json:"pg_code"`
		PgName string `json:"pg_name"`
	} `json:"payment_channel"`
	ResponseCode string `json:"response_code"`
	ResponseDesc string `json:"response_desc"`
}
