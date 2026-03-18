package security_cameras

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go-api/helpers"
	"go-api/modules/models"
)

type controller struct {
	service Service
}

func (c controller) GetDetailAction(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, _ := strconv.ParseInt(idParam, 10, 64)

	detail, isExists, err := c.service.GetDetail(ctx, id)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	} else if !isExists {
		helpers.NewResponse(ctx, http.StatusNotFound, "data not found")
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, detail)
	return
}

func (c controller) GetListAction(ctx *gin.Context) {
	pageQuery := ctx.DefaultQuery("page", "1")
	page, _ := strconv.ParseInt(pageQuery, 10, 64)

	limitQuery := ctx.DefaultQuery("limit", "10")
	limit, _ := strconv.ParseInt(limitQuery, 10, 64)

	filter := RequestListFilter{
		Page:    int(page),
		Limit:   int(limit),
		OrderBy: ctx.DefaultQuery("order_by", ""),
	}

	items, paginator, err := c.service.GetList(ctx, filter)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	result := map[string]interface{}{
		"paginator": paginator,
		"items":     items,
	}

	helpers.NewResponse(ctx, http.StatusOK, result)
	return
}

func (c controller) CreateAction(ctx *gin.Context) {
	var input models.SecurityCameras
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	result, err := c.service.Create(ctx, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, result)
	return
}

func (c controller) UpdateAction(ctx *gin.Context) {
	var input models.SecurityCameras
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	result, err := c.service.Update(ctx, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, result)
	return
}

func (c controller) DeleteAction(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, _ := strconv.ParseInt(idParam, 10, 64)

	err := c.service.Delete(ctx, id)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "Success Delete")
	return
}
