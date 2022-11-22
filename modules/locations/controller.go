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

func (c controller) GetDetailAction(ctx *gin.Context) {
	type UrlParams struct {
		LocationID string `uri:"locationID" binding:"required,uuid"`
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

func (c controller) UpdateAction(ctx *gin.Context) {
	var input RequestUpdateLocations
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	type UrlParams struct {
		LocationID string `uri:"locationID" binding:"required,uuid"`
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

func (c controller) UpdateStatusAction(ctx *gin.Context) {
	type UrlParams struct {
		LocationID string `uri:"locationID" binding:"required,uuid"`
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

func (c controller) DeleteAction(ctx *gin.Context) {
	type UrlParams struct {
		LocationID string `uri:"locationID" binding:"required,uuid"`
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

func (c controller) GetDetailActionLevel(ctx *gin.Context) {
	type UrlParams struct {
		LevelID string `uri:"levelID" binding:"required,uuid"`
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

func (c controller) UpdateActionLevel(ctx *gin.Context) {
	var input RequestUpdateLocationLevel
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	type UrlParams struct {
		LevelID string `uri:"levelID" binding:"required,uuid"`
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

func (c controller) UpdateStatusActionLevel(ctx *gin.Context) {
	type UrlParams struct {
		LocationID string `uri:"locationID" binding:"required,uuid"`
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

func (c controller) DeleteActionLevel(ctx *gin.Context) {
	type UrlParams struct {
		LocationID string `uri:"locationID" binding:"required,uuid"`
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
