package models

// M : Mandatory
// O : Optional
// C : Conditional

/*
PostDataTransactionRequest
| Parameter                     | Data Type                         | M/O/C | Description                                                                                                                   |
|-------------------- ----------|-----------------------------------|-------|-------------------------------------------------------------------------------------------------------------------------------|
| Request                       | Alfanumeric (50)                  | M     | Request Description                                                                                                           |
| Merchant_id                   | Numeric (5)                       | M     | Merchant Code From Faspay = BOI                                                                                               |
| Merchant                      | Alfanumeric (32)                  | M     | Merchant Name                                                                                                                 |
| bill_no                       | Alfanumeric (32)                  | M     | Order Number                                                                                                                  |
| bill_reff                     | Alfanumeric (32)                  | O     | Booking Number/reffrence (can fill same with order no)                                                                        |
| bill_date                     | Datetime YYYY-MM-DD HH:MM:SS (19) | M     | Transaction/ Order Date                                                                                                       |
| bill_expired                  | Datetime YYYY-MM-DD HH:MM:SS (19) | M     | Payment Expiring Date (max 30 days)                                                                                           |
| bill_desc                     | Alfanumeric (128)                 | M     | Transaction Description                                                                                                       |
| bill_currency                 | Alfanumeric (3)                   | M     | Currency, Must be 'IDR'                                                                                                       |
| bill_gross                    | Numeric (15)                      | O     | Order Nominal                                                                                                                 |
| bill_miscfee                  | Numeric (15)                      | O     | Miscellaneous fee                                                                                                             |
| bill_total                    | Numeric (15)                      | M     | Total Nominal                                                                                                                 |
| payment_channel               | Numeric (32)                      | M     | Payment Channel Code                                                                                                          |
| pay_type                      | Alfanumeric (1)                   | M     | Payment type : 1: Full Settlement 2: Installment  3: Mixed 1 &amp; 2 Pay Type 2 &amp; 3 only implement on BCA KlikPay channel |
| cust_no                       | Alfanumeric (32)                  | M     | Customer Number                                                                                                               |
| cust_name                     | Alfanumeric (128)                 | M     | Customer Name                                                                                                                 |
| bank_user_id                  | Alfanumeric (128)                 | O     | Customer User ID on bank’s services (ex : KlikBCA User Id)                                                                    |
| Msisdn                        | Numeric (64)                      | M     | Customer Mobile Phone                                                                                                         |
| Email                         | Alfanumeric (128)                 | M     | Customer Email                                                                                                                |
| Terminal                      | Numeric (16)                      | M     | Always use 10 for Terminal                                                                                                    |
| billing_name                  | Alfanumeric                       | C     | Billing name  for OVO                                                                                                         |
| billing_lastname              | Billing name                      | O     | biling last name                                                                                                              |
| billing_address               | Alfanumeric (200)                 | O     | billing_address                                                                                                               |
| billing_address_city          | Alfanumeric (50)                  | O     | Billing City                                                                                                                  |
| billing_address_region        | Alfanumeric (100)                 | O     | Billing Addres Region                                                                                                         |
| billing_address_state         | Alfanumeric (100)                 | O     | Billing Address State                                                                                                         |
| billing_address_poscode       | Alfanumeric (10)                  | O     | Billing Address Pos Code                                                                                                      |
| billing_address_country_code  | Alfanumeric (10)                  | O     | Billing Address Country Code                                                                                                  |
| receiver_name_for_shipping    | Alfanumeric (100)                 | C     | Receiver Name                                                                                                                 |
| shipping_lastname             | Alfanumeric                       | O     | Receiver last name                                                                                                            |
| shipping_address              | Alfanumeric (200)                 | O     | Shipping Address                                                                                                              |
| shipping_address_city         | Alfanumeric (50)                  | O     | Shipping Address City                                                                                                         |
| shipping_address_region       | Alfanumeric (100)                 | O     | Shipping Address Region                                                                                                       |
| shipping_address_state        | Alfanumeric (100)                 | O     | Shipping Address State                                                                                                        |
| shipping_address_poscode      | Alfanumeric (10)                  | O     | Shipping Address Pos Code                                                                                                     |
| shipping_address_country_code | Alfanumeric (10)                  | O     | Shipping Address Country Code                                                                                                 |
| shipping_msisdn               | Numeric                           | C     | Shipping phone number                                                                                                         |
| Product                       | Alfanumeric (50)                  | M     | Item Name                                                                                                                     |
| Amount                        | Numeric                           | M     | Item Price                                                                                                                    |
| Qty                           | Numeric (32)                      | M     | Item Quantity                                                                                                                 |
| payment_plan                  | Numeric (1)                       | M     | Payment plan 1: Full Settlement 2: Installement                                                                               |
| tenor                         | Numeric (2)                       | M     | Installment Tenor 00: Full Payment 03: 3 months 06: 6 months 12: 12 months Tenor 03,06,12 only use on BCA KlikPay channel     |
| merchant_id                   | Numeric (5)                       | M     | Merchant Id From Payment Channel ex : MID from BCA KlikPay                                                                    |
| Reserve1                      | Alfanumeric (50)                  | O     |                                                                                                                               |
| Reserve2                      | Alfanumeric (50)                  | C     |                                                                                                                               |
| Signature                     | Alfanumeric (100)                 | M     | sha1(md5(user_id merchant + password merchant + bill_no))                                                                     |

Note:
- Writing without separators, like BillTotal, BillGross, BillMiscfee, etc. Example: Rp. 10.000,00 Written 1000000
- Signature for Post Data : $signature = sha1(md5(($user_id.$pass.$bill_no)));

Example PostDataTransactionRequest:
{
  "request":"Post Data Transaction",
  "merchant_id":"98765",
  "merchant":"FASPAY",
  "bill_no":"98765123456789",
  "bill_reff":"12345678",
  "bill_date":"2020-09-02 10:48:10",
  "bill_expired":"2020-09-03 10:48:10",
  "bill_desc":"Pembayaran #12345678",
  "bill_currency":"IDR",
  "bill_gross":"0",
  "bill_miscfee":"0",
  "bill_total":"1000000",
  "cust_no":"12",
  "cust_name":"Nur Auliya",
  "payment_channel":"302",
  "pay_type":"1",
  "bank_userid":"",
  "msisdn":"628122131187",
  "email":"faspay@gmail.com",
  "terminal":"10",
  "billing_name":"0",
  "billing_lastname":"0",
  "billing_address":"jalan pintu air raya",
  "billing_address_city":"Jakarta Pusat",
  "billing_address_region":"DKI Jakarta",
  "billing_address_state":"Indonesia",
  "billing_address_poscode":"10710",
  "billing_msisdn":"",
  "billing_address_country_code":"ID",
  "receiver_name_for_shipping":"Nur Auliya",
  "shipping_lastname":"",
  "shipping_address":"jalan pintu air raya",
  "shipping_address_city":"Jakarta Pusat",
  "shipping_address_region":"DKI Jakarta",
  "shipping_address_state":"Indonesia",
  "shipping_address_poscode":"10710",
  "shipping_msisdn":"",
  "shipping_address_country_code":"ID",
  "item":[
    {
      "product":"Invoice No. inv-985/2017-03/1234567891",
      "qty":"1",
      "amount":"1000000",
      "payment_plan":"01",
      "merchant_id":"99999",
      "tenor":"00"
    }
  ],
  "reserve1":"",
  "reserve2":"",
  "signature":"5807a17ccd950904ec0a303725fa8a4b36c89e2f"
}
*/
type PostDataTransactionRequest struct {
	Request                    string `json:"request"`
	MerchantID                 string `json:"merchant_id"`
	Merchant                   string `json:"merchant"`
	BillNo                     string `json:"bill_no"`
	BillReff                   string `json:"bill_reff"`
	BillDate                   string `json:"bill_date"`
	BillExpired                string `json:"bill_expired"`
	BillDesc                   string `json:"bill_desc"`
	BillCurrency               string `json:"bill_currency"`
	BillGross                  string `json:"bill_gross"`
	BillMiscfee                string `json:"bill_miscfee"`
	BillTotal                  string `json:"bill_total"`
	CustNo                     string `json:"cust_no"`
	CustName                   string `json:"cust_name"`
	PaymentChannel             string `json:"payment_channel"`
	PayType                    string `json:"pay_type"`
	BankUserid                 string `json:"bank_userid"`
	Msisdn                     string `json:"msisdn"`
	Email                      string `json:"email"`
	Terminal                   string `json:"terminal"`
	BillingName                string `json:"billing_name"`
	BillingLastname            string `json:"billing_lastname"`
	BillingAddress             string `json:"billing_address"`
	BillingAddressCity         string `json:"billing_address_city"`
	BillingAddressRegion       string `json:"billing_address_region"`
	BillingAddressState        string `json:"billing_address_state"`
	BillingAddressPoscode      string `json:"billing_address_poscode"`
	BillingMsisdn              string `json:"billing_msisdn"`
	BillingAddressCountryCode  string `json:"billing_address_country_code"`
	ReceiverNameForShipping    string `json:"receiver_name_for_shipping"`
	ShippingLastname           string `json:"shipping_lastname"`
	ShippingAddress            string `json:"shipping_address"`
	ShippingAddressCity        string `json:"shipping_address_city"`
	ShippingAddressRegion      string `json:"shipping_address_region"`
	ShippingAddressState       string `json:"shipping_address_state"`
	ShippingAddressPoscode     string `json:"shipping_address_poscode"`
	ShippingMsisdn             string `json:"shipping_msisdn"`
	ShippingAddressCountryCode string `json:"shipping_address_country_code"`
	Item                       []struct {
		Product     string `json:"product"`
		Qty         string `json:"qty"`
		Amount      string `json:"amount"`
		PaymentPlan string `json:"payment_plan"`
		MerchantID  string `json:"merchant_id"`
		Tenor       string `json:"tenor"`
	} `json:"item"`
	Reserve1  string `json:"reserve1"`
	Reserve2  string `json:"reserve2"`
	Signature string `json:"signature"`
}

/*
PostDataTransactionResponse
| Parameter     | Data Type        | M/O/C | Description                                                                                                                                              |
|---------------|------------------|-------|----------------------------------------------------------------------------------------------------------------------------------------------------------|
| Response      | Alfanumeric (50) | O     | Response Name                                                                                                                                            |
| trx_id        | Numeric (16)     | M     | Transaction ID (Issued/generated by Faspay (Media Indonesia)) Notes: Unique Transaction ID for 1 day or as long as it hasn't been paid and hasn't expired |
| Merchant_id   | Numeric (5)      | M     | Merchant Code                                                                                                                                            |
| Merchant      | Alfanumeric (32) | M     | Merchant Name                                                                                                                                            |
| bill_no       | Alfanumeric (32) | M     | Order No                                                                                                                                                 |
| Response_Code | Numeric (2)      | M     | Response Code 00 Success                                                                                                                                 |
| Response_Desc | Alfanumeric (32) | M     | Response Code Description                                                                                                                                |
| redirect_url  | Alfanumeric      | O     | the redirect url for the next process, available only on JSON format                                                                                     |

Example PostDataTransactionResponse:
{
    "response": "Transmission Detail Info",
    "trx_id": "9876530200004184",
    "merchant_id": "98765",
    "merchant": "Faspay sandbox",
    "bill_no": "98765123456789",
    "bill_items": [
        {
            "product": "Invoice No. inv-985/2017-03/1234567891",
            "qty": "1",
            "amount": "1000000",
            "payment_plan": "01",
            "merchant_id": "99999",
            "tenor": "00"
        }
    ],
    "response_code": "00",
    "response_desc": "Sukses",
    "redirect_url": "https://dev.faspay.co.id/pws/100003/0830000010100000/5807a17ccd950904ec0a303725fa8a4b36c89e2f?trx_id=9876530200004184&merchant_id=98765&bill_no=98765123456789"
}
*/
type PostDataTransactionResponse struct {
	Response   string `json:"response"`
	TrxID      string `json:"trx_id"`
	MerchantID string `json:"merchant_id"`
	Merchant   string `json:"merchant"`
	BillNo     string `json:"bill_no"`
	BillItems  []struct {
		Product     string `json:"product"`
		Qty         string `json:"qty"`
		Amount      string `json:"amount"`
		PaymentPlan string `json:"payment_plan"`
		MerchantID  string `json:"merchant_id"`
		Tenor       string `json:"tenor"`
	} `json:"bill_items"`
	ResponseCode string `json:"response_code"`
	ResponseDesc string `json:"response_desc"`
	RedirectURL  string `json:"redirect_url"`
}
