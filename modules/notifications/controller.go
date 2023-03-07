package notifications

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"

	"go-api/helpers"
)

type controller struct {
	service Service
}

func (c controller) GetQRCodeWhatsapp(ctx *gin.Context) {
	qrCode, isExists, err := c.service.QRCodeWhatsapp(ctx)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"message":       "error getting whatsapp qr code",
			"error_message": err.Error(),
		})
		return
	} else if !isExists {
		helpers.NewResponse(ctx, http.StatusNotFound, gin.H{
			"message": "qr code not found, please start whatsapp notification task, and try again",
		})
		return
	}

	imageQrCode, err := qrcode.Encode(qrCode, qrcode.Medium, 256)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error_message": err.Error(),
			"message":       "error create qrcode",
		})
		return
	}

	buf := bytes.NewBuffer(imageQrCode)

	responseWriter := ctx.Writer
	responseWriter.Header().Set("Content-Type", "image/png")
	responseWriter.WriteHeader(http.StatusOK)
	_, _ = io.Copy(responseWriter, buf)
	return
}

func (c controller) GetDetailAction(ctx *gin.Context) {
	helpers.NewResponse(ctx, http.StatusOK, "Yey")
	return
}

func (c controller) GetListAction(ctx *gin.Context) {
	helpers.NewResponse(ctx, http.StatusOK, "Yey")
	return
}

func (c controller) CreateAction(ctx *gin.Context) {
	helpers.NewResponse(ctx, http.StatusOK, "Yey")
	return
}

func (c controller) UpdateAction(ctx *gin.Context) {
	helpers.NewResponse(ctx, http.StatusOK, "Yey")
	return
}

func (c controller) DeleteAction(ctx *gin.Context) {
	helpers.NewResponse(ctx, http.StatusOK, "Yey")
	return
}
