package locations

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"go-api/constant"
	"go-api/helpers"
)

type controller struct {
	service Service
}

// GetAllAction godoc
// @Summary      List locations
// @Tags         locations
// @Produce      json
// @Param        code  query  string  false  "Code"
// @Param        name  query  string  false  "Name"
// @Param        order  query  string  false  "asc|desc"
// @Param        order_by  query  string  false  "Order by (default name)"
// @Success      200  {object}  helpers.Response
// @Failure      400  {object}  helpers.Response
// @Router       /locations/ [get]
func (c controller) GetAllAction(ctx *gin.Context) {
	filter := FilterQueryLocations{
		Code:    ctx.DefaultQuery("code", ""),
		Name:    ctx.DefaultQuery("name", ""),
		Order:   ctx.DefaultQuery("order", "asc"),
		OrderBy: ctx.DefaultQuery("order_by", "name"),
	}

	locationModels, err := c.service.GetAllLocation(ctx, filter)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to get detail location, please try again",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, locationModels)
	return
}

// GetDetailAction godoc
// @Summary      Get location detail
// @Tags         locations
// @Produce      json
// @Param        locationID  path  string  true  "Location ID"
// @Success      200  {object}  helpers.Response
// @Failure      400  {object}  helpers.Response
// @Failure      404  {object}  helpers.Response
// @Router       /locations/{locationID} [get]
func (c controller) GetDetailAction(ctx *gin.Context) {
	type UrlParams struct {
		LocationID uint64 `uri:"locationID" binding:"required,uuid"`
	}

	var urlParams UrlParams
	if err := ctx.ShouldBindUri(&urlParams); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to get url params, please try again",
		})
		return
	}

	locationModel, notFound, err := c.service.GetDetailLocation(ctx, urlParams.LocationID)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to get detail location, please try again",
		})
		return
	} else if notFound {
		helpers.NewResponse(ctx, http.StatusNotFound, gin.H{
			"error_message": "location not found",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, locationModel)
	return
}

// CreateAction godoc
// @Summary      Create location
// @Tags         locations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  RequestCreateLocations  true  "Request body"
// @Success      200  {object}  helpers.Response
// @Failure      400  {object}  helpers.Response
// @Router       /locations/create [post]
func (c controller) CreateAction(ctx *gin.Context) {
	var input RequestCreateLocations
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	_, err = c.service.CreateLocation(ctx, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to create location, please try again",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "Location successfully created")
	return
}

// UpdateAction godoc
// @Summary      Update location
// @Tags         locations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        locationID  path  string  true  "Location ID"
// @Param        body  body  RequestUpdateLocations  true  "Request body"
// @Success      200  {object}  helpers.Response
// @Failure      400  {object}  helpers.Response
// @Router       /locations/{locationID} [put]
func (c controller) UpdateAction(ctx *gin.Context) {
	var input RequestUpdateLocations
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	type UrlParams struct {
		LocationID uint64 `uri:"locationID" binding:"required,uuid"`
	}

	var urlParams UrlParams
	if err = ctx.ShouldBindUri(&urlParams); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to get url params, please try again",
		})
		return
	}

	_, err = c.service.UpdateLocation(ctx, urlParams.LocationID, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to update location, please try again",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "Location successfully updated")
	return
}

// UpdateStatusAction godoc
// @Summary      Update location status
// @Tags         locations
// @Produce      json
// @Security     BearerAuth
// @Param        locationID  path  string  true  "Location ID"
// @Param        status  path  int  true  "Status"
// @Success      200  {object}  helpers.Response
// @Failure      400  {object}  helpers.Response
// @Router       /locations/{locationID}/status/{status} [put]
func (c controller) UpdateStatusAction(ctx *gin.Context) {
	type UrlParams struct {
		LocationID uint64 `uri:"locationID" binding:"required,uuid"`
		Status     string `uri:"status" binding:"required"`
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

	_, err = c.service.UpdateStatusLocation(ctx, urlParams.LocationID, statusInt)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to update status location, please try again",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "Location successfully updated status")
	return
}

// DeleteAction godoc
// @Summary      Delete location
// @Tags         locations
// @Produce      json
// @Security     BearerAuth
// @Param        locationID  path  string  true  "Location ID"
// @Success      200  {object}  helpers.Response
// @Failure      400  {object}  helpers.Response
// @Router       /locations/{locationID} [delete]
func (c controller) DeleteAction(ctx *gin.Context) {
	type UrlParams struct {
		LocationID uint64 `uri:"locationID" binding:"required,uuid"`
	}

	var urlParams UrlParams
	if err := ctx.ShouldBindUri(&urlParams); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to get url params, please try again",
		})
		return
	}

	err := c.service.DeleteLocation(ctx, urlParams.LocationID)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to delete location, please try again",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "Location successfully deleted")
	return
}

// -------------------------------

// GetAllActionLevel godoc
// @Summary      List location levels
// @Tags         location-levels
// @Produce      json
// @Security     BearerAuth
// @Param        name  query  string  false  "Name"
// @Param        order  query  string  false  "asc|desc"
// @Param        order_by  query  string  false  "Order by (default name)"
// @Success      200  {object}  helpers.Response
// @Failure      400  {object}  helpers.Response
// @Router       /locations/levels/ [get]
func (c controller) GetAllActionLevel(ctx *gin.Context) {
	filter := FilterQueryLocationLevel{
		Name:    ctx.DefaultQuery("name", ""),
		Order:   ctx.DefaultQuery("order", "asc"),
		OrderBy: ctx.DefaultQuery("order_by", "name"),
	}

	locationModels, err := c.service.GetAllLocationLevel(ctx, filter)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to get detail location level, please try again",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, locationModels)
	return
}

// GetDetailActionLevel godoc
// @Summary      Get location level detail
// @Tags         location-levels
// @Produce      json
// @Security     BearerAuth
// @Param        levelID  path  string  true  "Level ID"
// @Success      200  {object}  helpers.Response
// @Failure      400  {object}  helpers.Response
// @Failure      404  {object}  helpers.Response
// @Router       /locations/levels/{levelID} [get]
func (c controller) GetDetailActionLevel(ctx *gin.Context) {
	type UrlParams struct {
		LevelID uint64 `uri:"levelID" binding:"required,uuid"`
	}

	var urlParams UrlParams
	if err := ctx.ShouldBindUri(&urlParams); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to get url params, please try again",
		})
		return
	}

	locationLevelModel, notFound, err := c.service.GetDetailLocationLevel(ctx, urlParams.LevelID)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to get detail location level, please try again",
		})
		return
	} else if notFound {
		helpers.NewResponse(ctx, http.StatusNotFound, gin.H{
			"error_message": "location level not found",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, locationLevelModel)
	return
}

// CreateActionLevel godoc
// @Summary      Create location level
// @Tags         location-levels
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  RequestCreateLocationLevel  true  "Request body"
// @Success      200  {object}  helpers.Response
// @Failure      400  {object}  helpers.Response
// @Router       /locations/levels/create [post]
func (c controller) CreateActionLevel(ctx *gin.Context) {
	var input RequestCreateLocationLevel
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	_, err = c.service.CreateLocationLevel(ctx, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to create location level, please try again",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "Location level successfully created")
	return
}

// UpdateActionLevel godoc
// @Summary      Update location level
// @Tags         location-levels
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        levelID  path  string  true  "Level ID"
// @Param        body  body  RequestUpdateLocationLevel  true  "Request body"
// @Success      200  {object}  helpers.Response
// @Failure      400  {object}  helpers.Response
// @Router       /locations/levels/{levelID} [put]
func (c controller) UpdateActionLevel(ctx *gin.Context) {
	var input RequestUpdateLocationLevel
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	type UrlParams struct {
		LevelID uint64 `uri:"levelID" binding:"required,uuid"`
	}

	var urlParams UrlParams
	if err = ctx.ShouldBindUri(&urlParams); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to get url params, please try again",
		})
		return
	}

	_, err = c.service.UpdateLocationLevel(ctx, urlParams.LevelID, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to update location level, please try again",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "Location level successfully updated")
	return
}

// UpdateStatusActionLevel godoc
// @Summary      Update location level status
// @Tags         location-levels
// @Produce      json
// @Security     BearerAuth
// @Param        levelID  path  string  true  "Level ID"
// @Param        status  path  int  true  "Status"
// @Success      200  {object}  helpers.Response
// @Failure      400  {object}  helpers.Response
// @Router       /locations/levels/{levelID}/status/{status} [put]
func (c controller) UpdateStatusActionLevel(ctx *gin.Context) {
	type UrlParams struct {
		LocationID uint64 `uri:"locationID" binding:"required,uuid"`
		Status     string `uri:"status" binding:"required"`
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

	_, err = c.service.UpdateStatusLocation(ctx, urlParams.LocationID, statusInt)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to update status location, please try again",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "Location successfully updated status")
	return
}

// DeleteActionLevel godoc
// @Summary      Delete location level
// @Tags         location-levels
// @Produce      json
// @Security     BearerAuth
// @Param        levelID  path  string  true  "Level ID"
// @Success      200  {object}  helpers.Response
// @Failure      400  {object}  helpers.Response
// @Router       /locations/levels/{levelID}/delete [delete]
func (c controller) DeleteActionLevel(ctx *gin.Context) {
	type UrlParams struct {
		LocationID uint64 `uri:"locationID" binding:"required,uuid"`
	}

	var urlParams UrlParams
	if err := ctx.ShouldBindUri(&urlParams); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to get url params, please try again",
		})
		return
	}

	err := c.service.DeleteLocationLevel(ctx, urlParams.LocationID)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{
			"error":         err.Error(),
			"error_message": "failed to delete location, please try again",
		})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "Location successfully deleted")
	return
}
