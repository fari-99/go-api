package ws_auth

import (
	"errors"
	wsauth "github.com/fari-99/go-helper/ws_auth"
	"net/http"
	"strings"

	"github.com/fari-99/go-helper/token_generator"
	"github.com/gin-gonic/gin"

	"go-api/helpers"
)

type controller struct {
	service wsauth.Service
}

// IssueAction issues a WS token pair for the logged in (REST session) user.
// IssueAction godoc
// @Summary      Issue websocket token pair
// @Tags         ws-auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  helpers.Response
// @Failure      401  {object}  helpers.Response
// @Failure      429  {object}  helpers.Response
// @Failure      500  {object}  helpers.Response
// @Router       /storages/ws-token [post]
func (c controller) IssueAction(ctx *gin.Context) {
	uuid, _ := ctx.Get("uuid")
	uuidString, _ := uuid.(string)

	currentUser, err := helpers.GetCurrentUser(ctx, uuidString)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusUnauthorized, gin.H{"message": "You must login to access"})
		return
	}

	if !c.allowed(ctx, currentUser.ID.String()) {
		return
	}

	pair, err := c.service.Issue(ctx, ctx.Request, token_generator.UserDetails{
		ID:        currentUser.ID.String(),
		Email:     currentUser.Email,
		Username:  currentUser.Username,
		UserRoles: strings.Split(currentUser.Roles, ","),
	})
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{"message": "failed to issue token"})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, pair)
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshAction rotates a token pair. Authorization is the refresh token itself.
// RefreshAction godoc
// @Summary      Rotate websocket token pair
// @Tags         ws-auth
// @Accept       json
// @Produce      json
// @Param        body  body  refreshRequest  true  "Request body"
// @Success      200  {object}  helpers.Response
// @Failure      400  {object}  helpers.Response
// @Failure      401  {object}  helpers.Response
// @Failure      429  {object}  helpers.Response
// @Router       /storages/ws-token/refresh [post]
func (c controller) RefreshAction(ctx *gin.Context) {
	var req refreshRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, gin.H{"message": "refresh_token is required"})
		return
	}

	if !c.allowed(ctx, "ip:"+ctx.ClientIP()) {
		return
	}

	pair, _, err := c.service.Renew(ctx, ctx.Request, req.RefreshToken)
	if errors.Is(err, wsauth.ErrInvalidToken) || errors.Is(err, wsauth.ErrTokenReused) {
		helpers.NewResponse(ctx, http.StatusUnauthorized, gin.H{"message": "unauthorized"})
		return
	} else if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{"message": "failed to refresh token"})
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, pair)
}

func (c controller) allowed(ctx *gin.Context, key string) bool {
	ok, err := c.service.AllowTokenRequest(ctx, key)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{"message": "failed to process request"})
		return false
	}

	if !ok {
		helpers.NewResponse(ctx, http.StatusTooManyRequests, gin.H{"message": "too many requests, please try again later"})
		return false
	}

	return true
}
