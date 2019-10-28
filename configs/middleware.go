package configs

import (
	"errors"
	"go-api/helpers"
	"go-api/helpers/token_generator"
	"strings"

	"github.com/kataras/iris"

	"github.com/kataras/iris/context"
)

type MiddlewareConfiguration struct {
	AllowedRoles   []string `json:"allowed_roles"`
	AllowedAppName []string `json:"allowed_app_name"`
	Whitelist      []string `json:"whitelist"`
}

func DefaultConfig() MiddlewareConfiguration {
	config := MiddlewareConfiguration{
		AllowedRoles:   make([]string, 0, 2),
		AllowedAppName: make([]string, 0, 2),
		Whitelist:      make([]string, 0, 2),
	}

	return config
}

// checkRoles checks whether user roles are in allowed roles
func checkRoles(userRoles string, allowedRoles []string) bool {
	roles := strings.Split(userRoles, ",")
	found := false

	// Check if userRoles are in allowedRoles
	for _, role := range roles {
		isExist, _, _ := helpers.InArray(role, allowedRoles)
		if isExist {
			found = true
			break
		}
	}

	return found
}

// NewMiddleware inits auth middleware config and returns new handler
func NewMiddleware(config MiddlewareConfiguration) context.Handler {
	defaultConfig := DefaultConfig()

	// Assign allowed roles configuration
	if len(config.AllowedRoles) > 0 {
		defaultConfig.AllowedRoles = config.AllowedRoles
	}

	return defaultConfig.AuthServe
}

// AuthServe checks user data such as user ID and roles.
// If the data is valid, continues to next handler
func (config *MiddlewareConfiguration) AuthServe(ctx context.Context) {

	_, next, err := config.checkAuthHeader(ctx.GetHeader("Authorization"))
	if err != nil {
		_, _ = NewResponse(ctx, iris.StatusInternalServerError, iris.Map{
			"message":       "You must login to access",
			"error_message": err.Error(),
		})
		ctx.StopExecution()
		return
	}

	if !next {
		_, _ = NewResponse(ctx, iris.StatusInternalServerError, iris.Map{
			"message": "You must login to access",
		})
		ctx.StopExecution()
		return
	}

	// check app origin

	// check user roles

	// check whitelist

	ctx.Next()
}

func (config *MiddlewareConfiguration) checkAuthHeader(authHeader string) (*token_generator.JwtMapClaims, bool, error) {
	if len(authHeader) == 0 {
		return &token_generator.JwtMapClaims{}, false, errors.New("header Authorization Bearer Token is empty")
	}

	token := strings.Split(authHeader, " ")

	claims, err := token_generator.NewJwt().ParseToken(token[1])
	if err != nil {
		return &token_generator.JwtMapClaims{}, false, err
	}

	return claims, true, nil
}
