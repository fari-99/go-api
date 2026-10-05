package roles

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go-api/helpers"
)

type controller struct {
	service Service
}

// GetDetailAction godoc
// @Summary      Get role detail
// @Tags         roles
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Role ID"
// @Success      200  {object}  helpers.Response
// @Failure      400  {object}  helpers.Response
// @Failure      404  {object}  helpers.Response
// @Router       /roles/{id} [get]
func (c controller) GetDetailAction(ctx *gin.Context) {
	id := ctx.Param("id")
	detail, notFound, err := c.service.GetDetail(ctx, id)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	} else if notFound {
		helpers.NewResponse(ctx, http.StatusNotFound, "role not found")
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, detail)
	return
}

// CountPermissionsAction godoc
// @Summary      Count permissions of role
// @Tags         roles
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Role ID"
// @Success      200  {object}  helpers.Response
// @Failure      400  {object}  helpers.Response
// @Router       /roles/{id}/permissions-count [get]
func (c controller) CountPermissionsAction(ctx *gin.Context) {
	id := ctx.Param("id")
	count, err := c.service.CountPermissions(ctx, id)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to count permissions for role",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, gin.H{"permission_count": count})
	return
}

// GetListAction godoc
// @Summary      List roles
// @Tags         roles
// @Produce      json
// @Security     BearerAuth
// @Param        page  query  int  false  "Page (default 1)"
// @Param        limit  query  int  false  "Page size (default 10)"
// @Param        order_by  query  string  false  "Order by"
// @Success      200  {object}  helpers.Response
// @Failure      500  {object}  helpers.Response
// @Router       /roles/ [get]
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

// CreateAction godoc
// @Summary      Create role
// @Tags         roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  RequestRole  true  "Request body"
// @Success      200  {object}  helpers.Response
// @Failure      400  {object}  helpers.Response
// @Router       /roles/ [post]
func (c controller) CreateAction(ctx *gin.Context) {
	var input RequestRole
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	result, err := c.service.Create(ctx, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to create role",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, result)
	return
}

// UpdateAction godoc
// @Summary      Update role
// @Tags         roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Role ID"
// @Param        body  body  RequestRole  true  "Request body"
// @Success      200  {object}  helpers.Response
// @Failure      400  {object}  helpers.Response
// @Router       /roles/{id} [put]
func (c controller) UpdateAction(ctx *gin.Context) {
	id := ctx.Param("id")

	var input RequestRole
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	result, err := c.service.Update(ctx, id, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to update role",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, result)
	return
}

// DeleteAction godoc
// @Summary      Delete role
// @Tags         roles
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Role ID"
// @Success      200  {object}  helpers.Response
// @Failure      400  {object}  helpers.Response
// @Router       /roles/{id} [delete]
func (c controller) DeleteAction(ctx *gin.Context) {
	id := ctx.Param("id")
	err := c.service.Delete(ctx, id)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to delete role",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "Role successfully deleted")
	return
}
