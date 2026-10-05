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

// GetDetailAction godoc
// @Summary      Get security camera detail
// @Tags         security-cameras
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Camera ID"
// @Success      200  {object}  helpers.Response
// @Failure      400  {object}  helpers.Response
// @Failure      404  {object}  helpers.Response
// @Router       /security-cameras/{id} [get]
func (c controller) GetDetailAction(ctx *gin.Context) {
	id := ctx.Param("id")
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

// GetListAction godoc
// @Summary      List security cameras
// @Tags         security-cameras
// @Produce      json
// @Security     BearerAuth
// @Param        page  query  int  false  "Page (default 1)"
// @Param        limit  query  int  false  "Page size (default 10)"
// @Param        order_by  query  string  false  "Order by"
// @Success      200  {object}  helpers.Response
// @Failure      500  {object}  helpers.Response
// @Router       /security-cameras/ [get]
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

// CreateAction godoc
// @Summary      Create security camera
// @Tags         security-cameras
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  models.SecurityCameras  true  "Request body"
// @Success      200  {object}  helpers.Response
// @Failure      400  {object}  helpers.Response
// @Router       /security-cameras/ [post]
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

// UpdateAction godoc
// @Summary      Update security camera
// @Tags         security-cameras
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Camera ID"
// @Param        body  body  models.SecurityCameras  true  "Request body"
// @Success      200  {object}  helpers.Response
// @Failure      400  {object}  helpers.Response
// @Router       /security-cameras/{id} [put]
func (c controller) UpdateAction(ctx *gin.Context) {
	id := ctx.Param("id")
	var input models.SecurityCameras
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	result, err := c.service.Update(ctx, id, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, result)
	return
}

// DeleteAction godoc
// @Summary      Delete security camera
// @Tags         security-cameras
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Camera ID"
// @Success      200  {object}  helpers.Response
// @Failure      400  {object}  helpers.Response
// @Router       /security-cameras/{id} [delete]
func (c controller) DeleteAction(ctx *gin.Context) {
	id := ctx.Param("id")
	err := c.service.Delete(ctx, id)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "Success Delete")
	return
}
