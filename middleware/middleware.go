package middleware

import (
	"errors"
	"fmt"
	"go-api/configs"
	"go-api/helpers"
	"go-api/helpers/token_generator"
	"strings"

	"github.com/kataras/iris/v12"

	"github.com/kataras/iris/v12/context"
)

type BaseMiddleware struct {
	AllowedRoles   []string `json:"allowed_roles"`
	AllowedAppName []string `json:"allowed_app_name"`
	Whitelist      []string `json:"whitelist"`
	Blacklist      []string `json:"blacklist"`
}

func DefaultConfig() BaseMiddleware {
	config := BaseMiddleware{
		AllowedRoles:   make([]string, 0, 2),
		AllowedAppName: make([]string, 0, 2),
		Whitelist:      make([]string, 0, 2),
		Blacklist:      make([]string, 0, 2),
	}

	return config
}

// NewMiddleware inits auth middleware config and returns new handler
func NewMiddleware(config BaseMiddleware) context.Handler {
	defaultConfig := DefaultConfig()

	// Assign allowed roles configuration
	if len(config.AllowedRoles) > 0 {
		defaultConfig.AllowedRoles = config.AllowedRoles
	}

	return defaultConfig.AuthServe
}

// AuthServe checks user data such as user ID and roles.
// If the data is valid, continues to next handler
func (config *BaseMiddleware) AuthServe(ctx iris.Context) {

	claims, next, err := config.checkAuthHeader(ctx.GetHeader("Authorization"))
	if err != nil {
		_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, iris.Map{
			"message":       "You must login to access",
			"error_message": err.Error(),
		})
		ctx.StopExecution()
		return
	}

	if !next || !claims.TokenData.Authorized {
		_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, iris.Map{
			"message": "You must login to access",
		})
		ctx.StopExecution()
		return
	}

	userDetails := claims.UserDetails

	// setup uuid for controller
	ctx.Values().Set("uuid", claims.Uuid)
	ctx.Values().Set("user_details", userDetails.ID)

	// check app origin
	if appExists, _, _ := helpers.InArray(claims.TokenData.AppData.AppName, config.AllowedAppName); !appExists && len(config.AllowedAppName) > 0 {
		_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, iris.Map{
			"message": fmt.Sprintf("This Application is not supported by our system, please contact admin, app name := %s", claims.TokenData.AppData.AppName),
		})
		ctx.StopExecution()
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
			_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, iris.Map{
				"message": fmt.Sprintf("You don't have any roles to access this page"),
			})
			ctx.StopExecution()
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
			_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, iris.Map{
				"message": fmt.Sprintf("You can't access this page because your IP address is not on our whitelist"),
			})
			ctx.StopExecution()
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
			_, _ = configs.NewResponse(ctx, iris.StatusInternalServerError, iris.Map{
				"message": fmt.Sprintf("You can't access this page because your IP address is not on our whitelist"),
			})
			ctx.StopExecution()
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
