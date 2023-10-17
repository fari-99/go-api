package xendit

import (
	"log"

	"github.com/gin-gonic/gin"
)

func NewXenditRoutes(app *gin.Engine) {
	log.Println("Setup Test Xendit Payment Gateway router")
	control := controller{}

	balance := app.Group("/xendit/balance")
	{
		balance.GET("/", control.GetBalance)
	}

	customers := app.Group("/xendit/customers")
	{
		customers.POST("/", control.CustomersCreate)
	}

}
