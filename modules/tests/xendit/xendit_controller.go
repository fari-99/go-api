package xendit

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/xendit/xendit-go"
	"github.com/xendit/xendit-go/balance"
	"github.com/xendit/xendit-go/client"
	"github.com/xendit/xendit-go/customer"

	"go-api/helpers"
	"go-api/modules/configs"
)

type controller struct {
	*configs.DI
}

type ErrorResponse struct {
	Message      string      `json:"message"`
	ErrorMessage string      `json:"error_message"`
	ErrorDetails interface{} `json:"error_details,omitempty"`
}

const prodXendit = "prod"

func GetClient(clientType string) *client.API {
	xenditApiKey := os.Getenv("XENDIT_API_KEY") // get api key
	if xenditApiKey == "" {
		panic("xendit key is empty, please add XENDIT_API_KEY to env")
	}

	return client.New(xenditApiKey)
}

func (c controller) GetBalance(ctx *gin.Context) {
	var input balance.GetParams
	input.AccountType = xendit.BalanceAccountTypeEnum(ctx.DefaultQuery("account_type", ""))
	input.ForUserID = ctx.DefaultQuery("for_user_id", "")

	xenCli := GetClient(prodXendit)
	balanceResponse, err := xenCli.Balance.Get(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to get balance",
			ErrorMessage: err.Error(),
		})

		return
	}

	helpers.NewResponse(ctx, http.StatusOK, balanceResponse)
	return
}

func (c controller) CustomersCreate(ctx *gin.Context) {
	var input customer.CreateCustomerParams
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	xenCli := GetClient(prodXendit)
	customerModel, err := xenCli.Customer.CreateCustomer(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to create customers",
			ErrorMessage: err.Error(),
		})

		return
	}

	helpers.NewResponse(ctx, http.StatusOK, customerModel)
	return
}

func (c controller) CustomersGet(ctx *gin.Context) {
	var input customer.GetCustomerByReferenceIDParams
	input.ReferenceID = ctx.DefaultQuery("reference_id", "")

	xenCli := GetClient(prodXendit)
	customerModel, err := xenCli.Customer.GetCustomerByReferenceID(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to create customers",
			ErrorMessage: err.Error(),
		})

		return
	}

	helpers.NewResponse(ctx, http.StatusOK, customerModel)
	return
}
