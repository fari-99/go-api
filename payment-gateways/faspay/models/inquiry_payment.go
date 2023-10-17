package models

// M : Mandatory
// O : Optional
// C : Conditional

/*
InquiryPaymentRequest
| Parameter   | Data Type        | M/O/C | Description                                                                                                                                              |
|-------------|------------------|-------|----------------------------------------------------------------------------------------------------------------------------------------------------------|
| Request     | Alfanumeric (50) | O     | Request Description                                                                                                                                      |
| trx_id      | Numeric (16)     | M     | Transaction ID (Issued/generated by Faspay (Media Indonesia)) Notes: Unique Transaction ID for 1 day or as long as it hasn't been paid and hasn't expired |
| Merchant_id | Numeric (5)      | M     | Merchant Code                                                                                                                                            |
| bill_no     | Alfanumeric (16) | M     | Order Number                                                                                                                                             |
| Signature   | Alfanumeric      | M     | sha1(md5((user_id + password + bill_no))                                                                                                                 |


Example InquiryPaymentRequest:
{
    "request":"Inquiry Status Payment",
    "trx_id":"9876540000001115",
    "merchant_id":"98765",
    "bill_no":"9988776655",
    "signature":"b03951a6051dfa90894eea48ccb964619e8b3474"
}
*/
type InquiryPaymentRequest struct {
	Request    string `json:"request"`
	TrxID      string `json:"trx_id"`
	MerchantID string `json:"merchant_id"`
	BillNo     string `json:"bill_no"`
	Signature  string `json:"signature"`
}

/*
InquiryPaymentResponse
| Parameter           | Data Type        | M/O/C | Description                                                                                                                                              |
|---------------------|------------------|-------|----------------------------------------------------------------------------------------------------------------------------------------------------------|
| Response            | Alfanumeric (50) | O     | Response Name                                                                                                                                            |
| trx_id              | Numeric (16)     | M     | Transaction ID (Issued/generated by Faspay (Media Indonesia)) Notes: Unique Transaction ID for 1 day or as long as it hasn't been paid and hasn't expired |
| Merchant_id         | Numeric (6)      | M     | Merchant Code                                                                                                                                            |
| Merchant            | Alfanumeric (32) | M     | Merchant name                                                                                                                                            |
| bill_no             | Alfanumeric(32)  | M     | Order number                                                                                                                                             |
| payment_reff        | Alfanumeric (16) | M     | Payment Reff (from Payment Channel)                                                                                                                      |
| payment_date        | Date             | M     | Payment Date (from Payment Channel)                                                                                                                      |
| payment_status_code | Numeric (1)      | M     | Status Code 0 Unprocessed 1 In Process 2 Payment Success 4 Payment Reserval 5 No bills found 8 Payment Cancelled 9 Unknown                               |
| payment_status_desc | Alfanumeric (32) | M     | Description Status                                                                                                                                       |
| response_code       | Alfanumeric (2)  | M     | Response Code 00 Success                                                                                                                                 |
| response_desc       | Alfanumeric (32) | M     | Response Description                                                                                                                                     |

Example InquiryPaymentResponse:
{
    "response": "Pengecekan Status Pembayaran",
    "trx_id": "8985999800000588",
    "merchant_id": "99998",
    "merchant": "Rizki Store",
    "bill_no": "8515009999999999",
    "payment_reff": "123",
    "payment_date": "2016-07-22 14:59:30",
    "payment_status_code": "2",
    "payment_status_desc": "Payment Sukses",
    "response_code": "00",
    "response_desc": "Sukses"
}
*/
type InquiryPaymentResponse struct {
	Response          string `json:"response"`
	TrxID             string `json:"trx_id"`
	MerchantID        string `json:"merchant_id"`
	Merchant          string `json:"merchant"`
	BillNo            string `json:"bill_no"`
	PaymentReff       string `json:"payment_reff"`
	PaymentDate       string `json:"payment_date"`
	PaymentStatusCode string `json:"payment_status_code"`
	PaymentStatusDesc string `json:"payment_status_desc"`
	ResponseCode      string `json:"response_code"`
	ResponseDesc      string `json:"response_desc"`
}
