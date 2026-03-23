package auths

import (
	"fmt"
	"net/http"
	"os"

	"go-api/helpers"

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

	authData, notFound, err := c.service.AuthenticateUser(ctx, input)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	} else if notFound {
		helpers.NewResponse(ctx, http.StatusNotFound, fmt.Sprintf("email not found"))
		return
	}

	tokenCompiled := map[string]interface{}{
		"total_login":  authData.TotalLogin,
		"access_token": authData.Token.AccessToken,
		// "refresh_token":  authData.Token.RefreshToken,
		"two_fa_enabled": false,
	}

	if authData.UserModel.TwoFaEnabled {
		tokenCompiled["two_fa_enabled"] = true
		tokenCompiled["two_fa_models"] = authData.UserModel.TwoFaModels
	}

	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    authData.Token.RefreshToken,
		Path:     "/",
		Domain:   os.Getenv("PROJECT_DOMAIN"),
		Expires:  authData.Token.RefreshExpiredAt,
		Secure:   false,
		HttpOnly: true,
	})

	helpers.NewResponse(ctx, http.StatusOK, tokenCompiled)
	return
}

func (c controller) RefreshSession(ctx *gin.Context) {
	authData, isExists, err := c.service.RefreshAuth(ctx)
	if !isExists {
		helpers.NewResponse(ctx, http.StatusUnauthorized, "you need to re-login")
		return
	} else if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	tokenCompiled := map[string]interface{}{
		"total_login":  authData.TotalLogin,
		"access_token": authData.Token.AccessToken,
		// "refresh_token":  authData.Token.RefreshToken,
		"two_fa_enabled": false,
	}

	if authData.UserModel.TwoFaEnabled {
		tokenCompiled["two_fa_enabled"] = true
		tokenCompiled["two_fa_models"] = authData.UserModel.TwoFaModels
	}

	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    authData.Token.RefreshToken,
		Path:     "/",
		Domain:   os.Getenv("PROJECT_DOMAIN"),
		Expires:  authData.Token.RefreshExpiredAt,
		Secure:   false,
		HttpOnly: true,
	})

	helpers.NewResponse(ctx, http.StatusOK, tokenCompiled)
	return
}

func (c controller) GetAllSession(ctx *gin.Context) {
	allDevices, err := c.service.AllSessions(ctx)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, allDevices)
	return
}

func (c controller) SignOutAction(ctx *gin.Context) {
	totalLogin, notFound, err := c.service.SignOutUser(ctx)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	} else if notFound {
		helpers.NewResponse(ctx, http.StatusNotFound, fmt.Sprintf("email not found"))
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, fmt.Sprintf("Success sign out, total login left: %d", totalLogin))
	return
}

func (c controller) DeleteSession(ctx *gin.Context) {
	uuid, isExist := ctx.GetQuery("uuid")
	if !isExist {
		helpers.NewResponse(ctx, http.StatusOK, gin.H{
			"message": "uuid not found",
		})
		return
	}

	_, isExist, err := c.service.DeleteSession(ctx, uuid)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	} else if !isExist {
		helpers.NewResponse(ctx, http.StatusNotFound, fmt.Sprintf("session not found"))
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, fmt.Sprintf("Success delete session"))
	return
}

func (c controller) DeleteAllSessionAction(ctx *gin.Context) {
	isExist, err := c.service.DeleteAllSession(ctx)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	} else if !isExist {
		helpers.NewResponse(ctx, http.StatusNotFound, fmt.Sprintf("failed to get all sessions"))
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, fmt.Sprintf("Success delete all session"))
	return
}
