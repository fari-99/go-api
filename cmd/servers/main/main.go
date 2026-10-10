package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"go-api/modules/auths"
	"go-api/modules/configs"
	"go-api/modules/hasura"
	"go-api/modules/locations"
	"go-api/modules/middleware"
	"go-api/modules/notifications"
	"go-api/modules/permissions"
	"go-api/modules/roles"
	"go-api/modules/security_cameras"
	"go-api/modules/state_machine"
	"go-api/modules/storages"
	"go-api/modules/tests/xendit"
	"go-api/modules/twoFA"
	"go-api/modules/users"
	"go-api/modules/whatsapp"
	"go-api/modules/ws_auth"

	"github.com/gin-gonic/gin"

	wsauth "github.com/fari-99/go-helper/ws_auth"

	_ "github.com/joho/godotenv/autoload"
)

// @title                      Go API
// @version                    1.0
// @description                Go API service
// @BasePath                   /
// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description                Type "Bearer" followed by the access token
func main() {
	// get parameter from cli
	var host, port string
	flag.StringVar(&host, "host", os.Getenv("APP_HOST"), "host of the service")
	flag.StringVar(&port, "port", os.Getenv("GO_API_PORT"), "port of the service")
	flag.Parse()

	// info version service
	fmt.Printf("Service: %s\nVersion: %s\nParams:\n-host: host of the service\n-port: port of the service\nFramework:\n", os.Getenv("APP_NAME"), os.Getenv("APP_VER"))

	if rVal := recover(); rVal != nil {
		fmt.Printf("Rval: %+v\n", rVal)
	}

	// Setup routes and run application
	app := configs.GetGinApplication()
	di := configs.DIInit()
	authentication := middleware.AuthMiddleware(middleware.BaseMiddleware{})
	refreshAuth := middleware.RefreshAuthMiddleware(middleware.BaseMiddleware{})
	// otpMiddleware := middleware.TOTPMiddlewareLogin()
	rbacMiddleware := middleware.PermissionMiddleware()

	// RBAC needs the session set by auth, so it is chained after it per module (not app.Use)
	authorized := func(ctx *gin.Context) {
		authentication(ctx)
		if ctx.IsAborted() {
			return
		}
		rbacMiddleware(ctx)
	}

	// versions := middleware.VersionMiddleware(map[string]bool{
	//	"v0": false,
	//	"v1": true,
	// })

	// CORS middleware
	app.Use(middleware.CORSMiddleware())

	auths.NewRegistrator(app.Group(""),
		auths.NewService(auths.NewRepository(di)), authentication, refreshAuth)

	state_machine.NewRegistrator(app.Group(""),
		state_machine.NewService(state_machine.NewRepository(di)),
		authorized)

	// WebSocket upload is opt-in (WS_UPLOAD_ENABLED); when off, no WS secrets or Redis are needed
	var wsAuth wsauth.Service
	if ws_auth.Enabled() {
		wsAuth = wsauth.NewService(configs.GetRedis(configs.REDIS_WS_AUTH_PREFIX), wsauth.ConfigFromEnv())
		ws_auth.NewRegistrator(app.Group(""), wsAuth, authorized)
	}

	storages.NewRegistrator(app.Group(""),
		storages.NewService(storages.NewRepository(di)),
		authorized, wsAuth)

	notifications.NewRegistrator(app.Group(""),
		notifications.NewService(notifications.NewRepository(di)),
		authorized)

	twoFA.NewRegistrator(app.Group(""),
		twoFA.NewService(twoFA.NewRepository(di)),
		authorized)

	users.NewRegistrator(app.Group(""),
		users.NewService(users.NewRepository(di)),
		authorized)

	locations.NewRegistrator(app.Group(""),
		locations.NewService(locations.NewRepository(di)),
		authorized)

	permissions.NewRegistrator(app.Group(""),
		permissions.NewService(permissions.NewRepository(di)),
		authorized)

	roles.NewRegistrator(app.Group(""),
		roles.NewService(roles.NewRepository(di)),
		authorized)

	hasura.NewRegistrator(app.Group(""),
		hasura.NewService(hasura.NewRepository(di)))

	security_cameras.NewRegistrator(app.Group(""),
		security_cameras.NewService(security_cameras.NewRepository(di)),
		authorized)

	whatsapp.NewRegistrator(app.Group(""), di, authorized)

	xendit.NewXenditRoutes(app)

	applicationRun := fmt.Sprintf("%s:%s", host, port)
	log.Printf("Run application on %s", applicationRun)
	_ = app.Run(applicationRun)
}
