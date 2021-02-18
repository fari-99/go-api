package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"go-api/configs"
	"net/http"
)

type TelegramController struct {
	DB    *gorm.DB
	Redis *redis.Client
}

type TelegramAuthentication struct {
	Username string `json:"username"`
}

func (controller *TelegramController) AuthenticateAction(ctx *gin.Context) {
	var input TelegramAuthentication
	err := ctx.BindJSON(&input)
	if err != nil {
		configs.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	return
}
