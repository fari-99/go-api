package test_controllers

import (
	"go-api/configs"

	"github.com/jinzhu/gorm"
	"github.com/kataras/iris/v12"
	"gopkg.in/gomail.v2"
)

type EmailsController struct {
	DB          *gorm.DB
	EmailDialer *gomail.Dialer
}

func (controller *EmailsController) SendEmailAction(ctx iris.Context) {
	dialer := controller.EmailDialer

	m := gomail.NewMessage()
	m.SetHeader("From", "alex@example.com")
	m.SetHeader("To", "bob@example.com", "cora@example.com")
	m.SetAddressHeader("Cc", "dan@example.com", "Dan")
	m.SetHeader("Subject", "Hello!")
	m.SetBody("text/html", "Hello <b>Bob</b> and <i>Cora</i>!")
	//m.Attach("/home/Alex/lolcat.jpg")

	if err := dialer.DialAndSend(m); err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusOK, err.Error())
		return
	}

	_, _ = configs.NewResponse(ctx, iris.StatusOK, "yee")
	return
}
