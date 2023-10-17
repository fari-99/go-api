package flip

import (
	"log"

	"github.com/gin-gonic/gin"
)

func NewFlipRoutes(app *gin.Engine) {
	log.Println("Setup Test Flip Payment Gateway router")
	control := controller{}

	general := app.Group("/flip/general")
	{
		general.GET("/is-maintenance", control.IsMaintenance)
		general.GET("/balance", control.GetBalance)
		general.GET("/bank-info", control.GetBankInfo)
		general.POST("/bank-inquiry", control.BankInquiry) // failed, don't know why
	}

	moneyTransfer := app.Group("/flip/money-transfer")
	{
		moneyTransfer.POST("/disbursements", control.CreateDisbursement)
		moneyTransfer.GET("/disbursements", control.GetAllDisbursement)
		moneyTransfer.GET("/disbursements/details", control.GetDetailDisbursement) // failed, don't know why
	}

	specialMoneyTransfer := app.Group("/flip/money-transfer/special")
	{
		specialMoneyTransfer.POST("/disbursements", control.CreateSpecialDisbursement)
		specialMoneyTransfer.GET("/countries", control.GetDisbursementCountryList)
		specialMoneyTransfer.GET("/cities", control.GetDisbursementCityList)
		specialMoneyTransfer.GET("/country-city", control.GetDisbursementCountyCityList)
	}

	agentMoneyTransfer := app.Group("/flip/money-transfer/agent")
	{
		agentMoneyTransfer.POST("/disbursements", control.CreateAgentDisbursement)
		agentMoneyTransfer.GET("/disbursements", control.ListAgentDisbursement)
		agentMoneyTransfer.GET("/disbursements/:transactionID", control.DetailAgentDisbursement)
	}

	agentVerification := app.Group("/flip/agents")
	{
		agentVerification.POST("/", control.CreateAgents)
		agentVerification.PUT("/:agentID", control.UpdateAgent)
		agentVerification.GET("/:agentID", control.GetAgent)
		agentVerification.PUT("/:agentID/identity", control.UploadAgentImage)
		agentVerification.POST("/:agentID/documents", control.UploadAgentDocuments)
		agentVerification.PUT("/:agentID/submit", control.SubmitAgent)
		agentVerification.PUT("/:agentID/repair", control.RepairAgent)
		agentVerification.PUT("/:agentID/repair/photo", control.RepairImage)
		agentVerification.PUT("/:agentID/repair/photo-selfie", control.RepairSelfieImage)
		agentVerification.GET("/countries", control.AgentCountryList)
		agentVerification.GET("/provinces", control.AgentProvinceList)
		agentVerification.GET("/cities", control.AgentCityList)
		agentVerification.GET("/districts", control.AgentDistrictList)
	}

	acceptPayment := app.Group("/flip/bills")
	{
		acceptPayment.POST("/", control.CreateBill)
		acceptPayment.PUT("/:billID", control.EditBill)
		acceptPayment.GET("/", control.GetBillList)
		acceptPayment.GET("/:billID", control.GetBillDetail)
		acceptPayment.GET("/:billID/payments", control.GetBillPayments)
		acceptPayment.GET("/payments", control.GetPaymentList)
		acceptPayment.PUT("/payments/:transactionID/confirm", control.ConfirmBillPayment) // not checked yet (on live prod only)
	}

	internationalTransfer := app.Group("/flip/money-transfer/international")
	{
		internationalTransfer.GET("/exchange-rates", control.GetExchangeRates)
		internationalTransfer.GET("/form-data", control.GetFormData)
		internationalTransfer.POST("/c2c-c2b", control.CreateIntTransferC2X)
		//internationalTransfer.POST("/b2c-b2b", control.CreateIntTransferB2X)       // NOT FOUND
		internationalTransfer.GET("/:transactionID", control.GetIntTransferDetail)
		internationalTransfer.GET("/", control.GetIntTransferList)
	}
}
