package users

import (
	"log"

	"github.com/gin-gonic/gin"
)

func NewRegistrator(app *gin.RouterGroup, service Service, authHandler gin.HandlerFunc) {
	log.Println("Setup User router")
	control := controller{service: service}

	userPrivate := app.Group("/users")
	{
		userPrivate.Use(authHandler)
		userPrivate.POST("/create", control.CreateAction)
		userPrivate.GET("/profile", control.UserProfileAction)
		userPrivate.PUT("/change-password", control.ChangePasswordAction)

		userPrivate.GET("/", control.GetListAction)
		userPrivate.GET("/:id", control.GetDetailAction)
		userPrivate.PUT("/:id", control.UpdateAction)
		userPrivate.DELETE("/:id", control.DeleteAction)
		userPrivate.GET("/:id/roles", control.GetUserRolesAction)
		userPrivate.PUT("/:id/roles", control.UpdateUserRolesAction)
	}

	userPublic := app.Group("/users")
	{
		userPublic.POST("/forgot-password", control.ForgotPasswordAction)
		userPublic.POST("/forgot-username", control.ForgotUsernameAction)
		userPublic.POST("/reset-password", control.ResetPasswordAction)
	}
}
