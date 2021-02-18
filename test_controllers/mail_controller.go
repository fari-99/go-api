package test_controllers

import (
	"github.com/gin-gonic/gin"
	"go-api/configs"
	"net/http"

	"github.com/jinzhu/gorm"
	"gopkg.in/gomail.v2"
)

type EmailsController struct {
	DB          *gorm.DB
	EmailDialer *gomail.Dialer
}

func (controller *EmailsController) SendEmailAction(ctx *gin.Context) {
	dialer := controller.EmailDialer

	m := gomail.NewMessage()
	m.SetHeader("From", "alex@example.com")
	m.SetHeader("To", "bob@example.com", "cora@example.com")
	m.SetAddressHeader("Cc", "dan@example.com", "Dan")
	m.SetHeader("Subject", "Hello!")
	m.SetBody("text/html", "Hello <b>Bob</b> and <i>Cora</i>!")
	//m.Attach("/home/Alex/lolcat.jpg")

	if err := dialer.DialAndSend(m); err != nil {
		configs.NewResponse(ctx, http.StatusOK, err.Error())
		return
	}

	configs.NewResponse(ctx, http.StatusOK, "yee")
	return
}
