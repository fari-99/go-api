package calendar_managements

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"go-api/constant"
	"go-api/helpers"
)

type controller struct {
	service Service
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
	var input CreateCalendarManagementRequest
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	model, err := c.service.Create(ctx, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to create calendar management data",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, model)
	return
}

func (c controller) UpdateAction(ctx *gin.Context) {
	type UrlParams struct {
		id string `uri:"locationID" binding:"required,uuid"`
	}

	var urlParams UrlParams
	if err := ctx.ShouldBindUri(&urlParams); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to get url params, please try again",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "Yey")
	return
}

func (c controller) UpdateStatusAction(ctx *gin.Context) {
	type UrlParams struct {
		id     string `uri:"locationID" binding:"required,uuid"`
		Status string `uri:"status" binding:"required"`
	}

	var urlParams UrlParams
	if err := ctx.ShouldBindUri(&urlParams); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to get url params, please try again",
		})
		return
	}

	statusInt, err := constant.GetStatus(urlParams.Status)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to get status update data, please try again",
		})
		return
	}

	err = c.service.UpdateStatus(ctx, urlParams.id, statusInt)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to create calendar management data",
		})
		return
	}

	return
}

func (c controller) DeleteAction(ctx *gin.Context) {
	helpers.NewResponse(ctx, http.StatusOK, "Yey")
	return
}

func (c controller) GetBusinessDayAction(ctx *gin.Context) {

}
