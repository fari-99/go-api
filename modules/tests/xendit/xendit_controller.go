package xendit

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/xendit/xendit-go/v7"
	"github.com/xendit/xendit-go/v7/customer"

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

func GetClient(clientType string) *xendit.APIClient {
	xenditApiKey := os.Getenv("XENDIT_API_KEY") // get api key
	if xenditApiKey == "" {
		panic("xendit key is empty, please add XENDIT_API_KEY to env")
	}

	return xendit.NewClient(xenditApiKey)
}

func (c controller) GetBalance(ctx *gin.Context) {
	xenCli := GetClient(prodXendit)
	balanceResponse, httpRes, err := xenCli.BalanceApi.GetBalance(ctx).
		AccountType(ctx.DefaultQuery("account_type", "")).
		ForUserId(ctx.DefaultQuery("for_user_id", "")).
		Execute()
	if err != nil {
		log.Printf("full http resp %+v", httpRes)
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
	var input customer.CustomerRequest
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	xenCli := GetClient(prodXendit)
	customerModel, httpRes, err := xenCli.CustomerApi.CreateCustomer(ctx).
		CustomerRequest(input).
		Execute()
	if err != nil {
		log.Printf("full http resp %+v", httpRes)
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
	xenCli := GetClient(prodXendit)
	customerModel, httpRes, err := xenCli.CustomerApi.GetCustomerByReferenceID(ctx).
		ReferenceId(ctx.DefaultQuery("reference_id", "")).
		Execute()
	if err != nil {
		log.Printf("full http resp %+v", httpRes)
		helpers.NewResponse(ctx, http.StatusInternalServerError, ErrorResponse{
			Message:      "failed to create customers",
			ErrorMessage: err.Error(),
		})

		return
	}

	helpers.NewResponse(ctx, http.StatusOK, customerModel)
	return
}
