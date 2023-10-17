package faspay

import "fmt"

const LinkAja = "302"
const BRIMocash = "400"
const BRIePay = "401"
const PermataVA = "402"
const KlikBCA = "404"
const BcaKlikPay = "405"
const MaybankVA = "408"
const UNIcount = "410"
const OctoClicks = "700"
const DanamonOB = "701"
const BcaVA = "702"
const BcaSakuku = "704"
const PaymentPointIndomaret = "706"
const AlfaGroup = "707"
const Kredivo = "709"
const ShopeePayQRIS = "711"
const ShopeePayApp = "713"
const LinkajaApp = "716"
const BriVA = "800"
const BniVA = "801"
const MandiriVA = "802"
const Akulaku = "807"
const BSecurePG = "810"
const Ovo = "812"
const Maybank2U = "814"
const SinarmasVA = "818"
const Dana = "819"
const Indodana = "820"
const CimbVA = "825"

func AllLabelPaymentChannel() map[string]string {
	mapPaymentChannel := map[string]string{
		LinkAja:               "LinkAja",
		BRIMocash:             "BRI Mocash",
		BRIePay:               "BRI e-Pay",
		PermataVA:             "Permata Virtual Account",
		KlikBCA:               "KlikBCA",
		BcaKlikPay:            "BCA KlikPay",
		MaybankVA:             "Maybank Virtual Account",
		UNIcount:              "UNIcount",
		OctoClicks:            "OctoClicks",
		DanamonOB:             "Danamon Online Banking",
		BcaVA:                 "BCA Virtual Account",
		BcaSakuku:             "BCA Sakuku",
		PaymentPointIndomaret: "Payment Point Indomaret",
		AlfaGroup:             "Alfagroup",
		Kredivo:               "Kredivo",
		ShopeePayQRIS:         "ShopeePay QRIS",
		ShopeePayApp:          "ShopeePay App",
		LinkajaApp:            "Linkaja App",
		BriVA:                 "BRI Virtual Account",
		BniVA:                 "BNI Virtual Account",
		MandiriVA:             "Mandiri Virtual Account",
		Akulaku:               "Akulaku",
		BSecurePG:             "B-Secure Payment Gateway",
		Ovo:                   "OVO",
		Maybank2U:             "Maybank2U",
		SinarmasVA:            "Sinarmas Virtual Account",
		Dana:                  "DANA",
		Indodana:              "Indodana",
		CimbVA:                "CIMB Virtual Account",
	}

	return mapPaymentChannel
}

func GetLabelPaymentChannel(paymentChannel string) (string, error) {
	mapPaymentChannel := AllLabelPaymentChannel()
	if _, ok := mapPaymentChannel[paymentChannel]; ok {
		return mapPaymentChannel[paymentChannel], nil
	}

	// https://docs..co.id/merchant-integration/api-reference-1/debit-transaction/reference/payment-channel-code
	return "", fmt.Errorf("payment channel not set, please check  documentation for any payment channle update")
}
