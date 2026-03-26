package whatsapp

import (
	"net/http"

	"go-api/constant"
	"go-api/modules/configs"

	"github.com/gin-gonic/gin"

	"go-api/helpers"
)

type controller struct {
	di *configs.DI
}

func (c controller) QRCodeAction(ctx *gin.Context) {
	redisClient := c.di.RedisSession

	qrCode, err := redisClient.Get(ctx, constant.QRCodeWhatsapp).Result()
	if err != nil {
		helpers.NewResponse(ctx, http.StatusNotFound, gin.H{
			"message": "QR code not available — either not initiated, already connected, or expired",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, gin.H{
		"message": "scan within 60 seconds",
		"qr_code": qrCode,
	})
	return
}

func (c controller) LoginAction(ctx *gin.Context) {
	redisClient := c.di.RedisSession

	err := configs.InitiateWhatsappLogin(ctx, redisClient)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, gin.H{
		"message": "QR generation triggered, poll GET /whatsapp/qr-code to retrieve it",
	})
	return
}
