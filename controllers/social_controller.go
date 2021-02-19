package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	"go-api/configs"
	"go-api/constant"
	"go-api/helpers"
	"go-api/models"
	"net/http"
	"os"
)

type TelegramController struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func (controller *TelegramController) AuthenticateAction(ctx *gin.Context) {
	userUuid, exists := ctx.Get("uuid")
	if !exists {
		configs.NewResponse(ctx, http.StatusOK, "Customer not login or authentication failed")
		return
	}

	customerModel, err := helpers.GetCurrentUser(userUuid.(string))
	if err != nil {
		configs.NewResponse(ctx, http.StatusOK, err.Error())
		return
	}

	telegram := models.SocialMedia{
		UserID: customerModel.ID,
		Name:   "telegram",
		Status: constant.StatusActive,
	}

	db := controller.DB
	err = db.Create(&telegram).Error
	if err != nil {
		configs.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	configs.NewResponse(ctx, http.StatusOK, gin.H{
		"message": "success create telegram",
		"url":     fmt.Sprintf("%s/%s?start=%s", os.Getenv("TELEGRAM_SITE"), os.Getenv("TELEGRAM_BOT_USERNAME"), telegram.Uuid),
	})
	return
}
