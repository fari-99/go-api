package middleware

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-api/configs"
	"go-api/helpers"
	"go-api/helpers/token_generator"
	"net/http"
	"strings"
)

// AuthMiddleware inits auth middleware config and returns new handler
func AuthMiddleware(config BaseMiddleware) gin.HandlerFunc {
	defaultConfig := DefaultConfig()

	// Assign allowed roles configuration
	if len(config.AllowedRoles) > 0 {
		defaultConfig.AllowedRoles = config.AllowedRoles
	}

	return defaultConfig.authServe
}

// authServe checks user data such as user ID and roles.
// If the data is valid, continues to next handler
func (config *BaseMiddleware) authServe(ctx *gin.Context) {
	cookie, err := ctx.Request.Cookie("token")

	var accessToken string
	if err != nil || err == http.ErrNoCookie {
		accessToken = ctx.GetHeader("Authorization")
	} else {
		accessToken = fmt.Sprintf("Bearer %s", cookie.Value)
	}

	claims, next, err := config.checkAuthHeader(accessToken)
	if err != nil {
		configs.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"message":       "You must login to access",
			"error_message": err.Error(),
		})
		ctx.Abort()
		return
	}

	if !next || !claims.TokenData.Authorized {
		configs.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"message": "You must login to access",
		})
		ctx.Abort()
		return
	}

	userDetails := claims.UserDetails

	// setup uuid for controller
	ctx.Set("uuid", claims.Uuid)
	ctx.Set("user_details", userDetails.ID)

	// check app origin
	if appExists, _, _ := helpers.InArray(claims.TokenData.AppData.AppName, config.AllowedAppName); !appExists && len(config.AllowedAppName) > 0 {
		configs.NewResponse(ctx, http.StatusInternalServerError, gin.H{
			"message": fmt.Sprintf("This Application is not supported by our system, please contact admin, app name := %s", claims.TokenData.AppData.AppName),
		})
		ctx.Abort()
		return
	}

	// check user roles
	if len(userDetails.UserRoles) > 0 && len(config.AllowedRoles) > 0 {
		var exists int
		for _, userRole := range userDetails.UserRoles {
			roleExists, _, _ := helpers.InArray(userRole, config.AllowedRoles)
			if roleExists {
				exists++
				break
			}
		}

		// no roles
		if exists == 0 {
			configs.NewResponse(ctx, http.StatusInternalServerError, gin.H{
				"message": fmt.Sprintf("You don't have any roles to access this page"),
			})
			ctx.Abort()
			return
		}
	}

	// check whitelist (all in the whitelist can access, otherwise can't)
	if len(claims.TokenData.AppData.IPList) > 0 && len(config.Whitelist) > 0 {
		var exists int
		for _, ipList := range claims.TokenData.AppData.IPList {
			ipListExist, _, _ := helpers.InArray(ipList, config.Whitelist)
			if ipListExist {
				exists++
				break
			}
		}

		// ip not on whitelist
		if exists == 0 {
			configs.NewResponse(ctx, http.StatusInternalServerError, gin.H{
				"message": fmt.Sprintf("You can't access this page because your IP address is not on our whitelist"),
			})
			ctx.Abort()
			return
		}
	}

	// check Blacklist (all in the blacklist can't access, otherwise can)
	if len(claims.TokenData.AppData.IPList) > 0 && len(config.Blacklist) > 0 {
		var exists int
		for _, ipList := range claims.TokenData.AppData.IPList {
			ipListExist, _, _ := helpers.InArray(ipList, config.Blacklist)
			if ipListExist {
				exists++
				break
			}
		}

		// ip on blacklist
		if exists > 0 {
			configs.NewResponse(ctx, http.StatusInternalServerError, gin.H{
				"message": fmt.Sprintf("You can't access this page because your IP address is not on our whitelist"),
			})
			ctx.Abort()
			return
		}
	}

	ctx.Next()
}

func (config *BaseMiddleware) checkAuthHeader(authHeader string) (*token_generator.JwtMapClaims, bool, error) {
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
