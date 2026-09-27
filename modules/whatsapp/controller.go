package whatsapp

import (
	"bytes"
	"io"
	"net/http"

	"go-api/constant"
	"go-api/modules/configs"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"

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

	imageQrCode, err := qrcode.Encode(qrCode, qrcode.Medium, 256)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"message":       "error creating qr code image",
			"error_message": err.Error(),
		})
		return
	}

	responseWriter := ctx.Writer
	responseWriter.Header().Set("Content-Type", "image/png")
	responseWriter.WriteHeader(http.StatusOK)
	_, _ = io.Copy(responseWriter, bytes.NewBuffer(imageQrCode))
	return
}

func (c controller) StatusAction(ctx *gin.Context) {
	redisClient := c.di.RedisSession

	status := configs.WhatsappConnectionStatus(ctx, redisClient)

	helpers.NewResponse(ctx, http.StatusOK, status)
	return
}

func (c controller) LogoutAction(ctx *gin.Context) {
	redisClient := c.di.RedisSession

	if err := configs.WhatsappLogout(ctx, redisClient); err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"message":       "error logging out of whatsapp",
			"error_message": err.Error(),
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, gin.H{
		"message": "device unlinked, call POST /whatsapp/login to re-pair",
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
