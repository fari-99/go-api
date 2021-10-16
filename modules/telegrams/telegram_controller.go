package telegrams

import (
	"github.com/gin-gonic/gin"
	"go-api/helpers"
	"go-api/modules/configs"
	"net/http"
)

type TelegramController struct {
	*configs.DI
}

type TelegramAuthentication struct {
	Username string `json:"username"`
}

func (controller *TelegramController) AuthenticateAction(ctx *gin.Context) {
	var input TelegramAuthentication
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	return
}
