package ws_auth

import (
	wsauth "github.com/fari-99/go-helper/ws_auth"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"go-api/helpers"
)

const (
	ContextSession         = "ws_session"
	ContextAccessExpiresAt = "ws_access_expires_at"
)

// Middleware decrypts the WS access token (Authorization Bearer header first, then ?token= because
// browsers can't set headers on a WebSocket) and validates it against JWT and the live Redis record.
// Every failure is a generic 401.
func Middleware(svc wsauth.Service) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := strings.TrimSpace(strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer "))
		if token == "" {
			token = ctx.Query("token")
		}

		session, expiresAt, err := svc.Authenticate(ctx.Request.Context(), token)
		if err != nil {
			helpers.NewResponse(ctx, http.StatusUnauthorized, gin.H{"message": "unauthorized"})
			ctx.Abort()
			return
		}

		ctx.Set(ContextSession, session)
		ctx.Set(ContextAccessExpiresAt, expiresAt)
		ctx.Next()
	}
}
