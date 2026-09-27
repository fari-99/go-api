package notifications

import (
	"bytes"
	"io"
	"net/http"
	"strconv"

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
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, "invalid notification template id")
		return
	}

	detail, notFound, err := c.service.GetDetail(ctx, id)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	} else if notFound {
		helpers.NewResponse(ctx, http.StatusNotFound, "notification template not found")
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, detail)
	return
}

// GetListAction lists notification templates. Filter by notification_type to get
// the templates available for a given channel (whatsapp, telegram, email, etc)
// when picking one to send manually to a user.
func (c controller) GetListAction(ctx *gin.Context) {
	pageQuery := ctx.DefaultQuery("page", "1")
	page, _ := strconv.ParseInt(pageQuery, 10, 64)

	limitQuery := ctx.DefaultQuery("limit", "10")
	limit, _ := strconv.ParseInt(limitQuery, 10, 64)

	notificationTypeQuery := ctx.DefaultQuery("notification_type", "0")
	notificationType, _ := strconv.ParseInt(notificationTypeQuery, 10, 8)

	filter := RequestListFilter{
		Page:             int(page),
		Limit:            int(limit),
		OrderBy:          ctx.DefaultQuery("order_by", ""),
		NotificationType: int8(notificationType),
		Action:           ctx.DefaultQuery("action", ""),
	}

	items, paginatorData, err := c.service.GetList(ctx, filter)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	result := map[string]interface{}{
		"paginator": paginatorData,
		"items":     items,
	}

	helpers.NewResponse(ctx, http.StatusOK, result)
	return
}

func (c controller) CreateAction(ctx *gin.Context) {
	var input RequestNotificationTemplate
	if err := ctx.BindJSON(&input); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := input.Validate(); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	model, err := c.service.Create(ctx, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"error":         err.Error(),
			"error_message": "failed to create notification template, please try again",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, model)
	return
}

func (c controller) UpdateAction(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, "invalid notification template id")
		return
	}

	var input RequestNotificationTemplate
	if err := ctx.BindJSON(&input); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := input.Validate(); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	model, err := c.service.Update(ctx, id, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"error":         err.Error(),
			"error_message": "failed to update notification template, please try again",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, model)
	return
}

func (c controller) DeleteAction(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, "invalid notification template id")
		return
	}

	if err := c.service.Delete(ctx, id); err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"error":         err.Error(),
			"error_message": "failed to delete notification template, please try again",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "notification template successfully deleted")
	return
}

// SendManualAction sends an existing notification template to a single user right away,
// dispatched through whichever channel (whatsapp, telegram, email, sms, push) the template is for.
func (c controller) SendManualAction(ctx *gin.Context) {
	var input RequestSendNotification
	if err := ctx.BindJSON(&input); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := input.Validate(); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if err := c.service.SendManual(ctx, input); err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"error":         err.Error(),
			"error_message": "failed to send notification, please try again",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "notification successfully sent")
	return
}
