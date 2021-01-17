package controllers

import (
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"
	"go-api/configs"
)

type TelegramController struct {
	DB    *gorm.DB
	Redis *sessions.Sessions
}

type TelegramAuthentication struct {
	Username string `json:"username"`
}

func (controller *TelegramController) AuthenticateAction(ctx iris.Context) {
	var input TelegramAuthentication
	err := ctx.ReadJSON(&input)
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusBadRequest, err.Error())
		return
	}

	return
}
