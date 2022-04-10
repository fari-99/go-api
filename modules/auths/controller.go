package auths

import (
	"fmt"
	"go-api/helpers"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type controller struct {
	service Service
}

func (c controller) AuthenticateAction(ctx *gin.Context) {
	var input RequestAuthUser
	err := ctx.BindJSON(&input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	token, notFound, err := c.service.AuthenticateUser(ctx, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	} else if notFound {
		helpers.NewResponse(ctx, http.StatusNotFound, fmt.Sprintf("email not found"))
		return
	}

	tokenCompiled := map[string]interface{}{
		"access_token":  token.AccessToken,
		"refresh_token": token.AccessToken,
	}

	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     "token",
		Value:    token.AccessToken,
		Path:     "/",
		Domain:   os.Getenv("PROJECT_DOMAIN"),
		Expires:  time.Unix(token.AccessExpiredAt, 0),
		Secure:   false,
		HttpOnly: true,
	})

	helpers.NewResponse(ctx, http.StatusOK, tokenCompiled)
	return
}
