package auths

import (
	"log"

	"github.com/gin-gonic/gin"
)

func NewRegistrator(app *gin.RouterGroup, service Service, authentication gin.HandlerFunc, refreshAuth gin.HandlerFunc) {
	log.Println("Setup Auth router")
	control := controller{service: service}

	userPublic := app.Group("/users")
	{
		// authentication data
		userPublic.POST("/auth", control.AuthenticateAction)
	}

	userPrivate := app.Group("/users/sessions")
	{
		userPrivate.Use(authentication)
		userPrivate.GET("/", control.GetAllSession)
		userPrivate.POST("/sign-out", control.SignOutAction)
		userPrivate.DELETE("/all", control.DeleteAllSessionAction)
		userPrivate.DELETE("/delete", control.DeleteSession)
	}

	userRefresh := app.Group("/users/sessions")
	{
		userRefresh.Use(refreshAuth)
		userRefresh.POST("/refresh", control.RefreshSession)
	}
}
