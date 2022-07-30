package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"go-api/helpers"
	"go-api/helpers/token_generator"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware inits auth middleware config and returns new handler
func AuthMiddleware(config BaseMiddleware) gin.HandlerFunc {
	defaultConfig := DefaultConfig()
	return defaultConfig.authServe
}

func RefreshAuthMiddleware(config BaseMiddleware) gin.HandlerFunc {
	defaultConfig := DefaultConfig()
	return defaultConfig.refreshServe
}

// authServe checks user data such as user ID and roles.
// If the data is valid, continues to next handler
func (base *BaseMiddleware) authServe(ctx *gin.Context) {
	cookie, err := ctx.Request.Cookie("access_token")

	var accessToken string
	if err != nil || err == http.ErrNoCookie {
		accessToken = ctx.GetHeader("Authorization")
	} else {
		accessToken = fmt.Sprintf("Bearer %s", cookie.Value)
	}

	claims, next, err := base.checkAuthHeader(accessToken)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"message":       "You must login to access",
			"error_message": err.Error(),
		})
		ctx.Abort()
		return
	}

	if !next || !claims.TokenData.Authorized {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"message": "You must login to access",
		})
		ctx.Abort()
		return
	}

	base.checkAuth(ctx, claims)
}

// authServe checks user data such as user ID and roles.
// If the data is valid, continues to next handler
func (base *BaseMiddleware) refreshServe(ctx *gin.Context) {
	cookie, err := ctx.Request.Cookie("refresh_token")

	var accessToken string
	if err != nil || err == http.ErrNoCookie {
		accessToken = ctx.GetHeader("Authorization")
	} else {
		accessToken = fmt.Sprintf("Bearer %s", cookie.Value)
	}

	claims, next, err := base.checkAuthHeader(accessToken)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"message":       "You must login to access",
			"error_message": err.Error(),
		})
		ctx.Abort()
		return
	}

	if !next || !claims.TokenData.Authorized {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"message": "You must login to access",
		})
		ctx.Abort()
		return
	}

	base.checkAuth(ctx, claims)
}

func (base *BaseMiddleware) checkAuth(ctx *gin.Context, claims *token_generator.JwtMapClaims) {
	userDetails := claims.UserDetails

	// setup uuid for controller
	ctx.Set("uuid", claims.Uuid)
	ctx.Set("user_details", userDetails.ID)

	_, err := helpers.GetCurrentUser(claims.Uuid)
	if err != nil {
		if isUsed, err := helpers.CheckFamily("", claims.Uuid); err != nil {
			helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			ctx.Abort()
			return
		} else if isUsed {
			helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
				"message": fmt.Sprintf("this token already used, please re-authenticate your account"),
			})
			ctx.Abort()
			return
		}

		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("authentication error, please re-login"),
		})
		ctx.Abort()
		return
	}

	// check app origin
	if appExists, _, _ := helpers.InArray(claims.TokenData.AppData.AppName, base.AllowedAppName); !appExists && len(base.AllowedAppName) > 0 {
		helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("This Application is not supported by our system, please contact admin, app name := %s", claims.TokenData.AppData.AppName),
		})
		ctx.Abort()
		return
	}

	// check whitelist (all in the whitelist can access, otherwise can't)
	if len(claims.TokenData.AppData.IPList) > 0 && len(base.Whitelist) > 0 {
		var exists int
		for _, ipList := range claims.TokenData.AppData.IPList {
			ipListExist, _, _ := helpers.InArray(ipList, base.Whitelist)
			if ipListExist {
				exists++
				break
			}
		}

		// ip not on whitelist
		if exists == 0 {
			helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
				"message": fmt.Sprintf("You can't access this page because your IP address is not on our whitelist"),
			})
			ctx.Abort()
			return
		}
	}

	// check Blacklist (all in the blacklist can't access, otherwise can)
	if len(claims.TokenData.AppData.IPList) > 0 && len(base.Blacklist) > 0 {
		var exists int
		for _, ipList := range claims.TokenData.AppData.IPList {
			ipListExist, _, _ := helpers.InArray(ipList, base.Blacklist)
			if ipListExist {
				exists++
				break
			}
		}

		// ip on blacklist
		if exists > 0 {
			helpers.NewResponse(ctx, http.StatusInternalServerError, gin.H{
				"message": fmt.Sprintf("You can't access this page because your IP address is not on our whitelist"),
			})
			ctx.Abort()
			return
		}
	}

	ctx.Next()
}

func (base *BaseMiddleware) checkAuthHeader(authHeader string) (*token_generator.JwtMapClaims, bool, error) {
	if len(authHeader) == 0 {
		return &token_generator.JwtMapClaims{}, false, errors.New("header Authorization Bearer Token is empty")
	}

	token := strings.Split(authHeader, " ")

	claims, err := token_generator.NewJwt().ParseToken(token[1])
	if err != nil {
		return &token_generator.JwtMapClaims{}, false, err
	}

	claims.TokenData.Authorized = true
	return claims, true, nil
}
