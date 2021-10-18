package telegrams

import (
	"github.com/gin-gonic/gin"
	"go-api/helpers"
	"net/http"
)

type controller struct {
	service Service
}

func (c controller) AuthenticateAction(ctx *gin.Context) {
	helpers.NewResponse(ctx, http.StatusOK, "yey")
	return
}
