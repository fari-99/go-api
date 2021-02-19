package test_controllers

import (
	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jinzhu/gorm"
	"go-api/configs"
	"net/http"
)

type TestSocialController struct {
	DB       *gorm.DB
	Telegram *tgbotapi.BotAPI
}

type TelegramTestInput struct {
	ChatID int64 `json:"chat_id"`
}

func (controller *TestSocialController) TestTelegramAction(ctx *gin.Context) {
	var input TelegramTestInput
	err := ctx.BindJSON(&input)
	if err != nil {
		configs.NewResponse(ctx, http.StatusOK, err.Error())
		return
	}

	bot := controller.Telegram

	msg := tgbotapi.NewMessage(input.ChatID, "testing message")
	_, _ = bot.Send(msg)

	configs.NewResponse(ctx, http.StatusOK, "yee")
	return
}
