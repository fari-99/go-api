package emails

import (
	"github.com/gin-gonic/gin"
	"go-api/helpers"
	"go-api/modules/configs"
	"net/http"

	"gopkg.in/gomail.v2"
)

type EmailsController struct {
	*configs.DI
}

func (controller *EmailsController) SendEmailAction(ctx *gin.Context) {
	dialer := controller.EmailDialler

	m := gomail.NewMessage()
	m.SetHeader("From", "alex@example.com")
	m.SetHeader("To", "bob@example.com", "cora@example.com")
	m.SetAddressHeader("Cc", "dan@example.com", "Dan")
	m.SetHeader("Subject", "Hello!")
	m.SetBody("text/html", "Hello <b>Bob</b> and <i>Cora</i>!")
	//m.Attach("/home/Alex/lolcat.jpg")

	if err := dialer.DialAndSend(m); err != nil {
		helpers.NewResponse(ctx, http.StatusOK, err.Error())
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "yee")
	return
}
