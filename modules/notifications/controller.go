package notifications

import (
	"github.com/gin-gonic/gin"
	"go-api/helpers"
	"net/http"
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
