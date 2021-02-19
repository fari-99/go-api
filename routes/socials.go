package routes

import (
	"go-api/controllers"
	"go-api/middleware"
	"log"
)

func (routes *Routes) setupSocialRoute() {
	log.Println("Setup Social router")

	app := routes.ginApp
	db := routes.DB
	redis := routes.Redis

	authentication := middleware.AuthMiddleware(middleware.BaseMiddleware{})

	// Approver Endpoint collection
	telegrams := app.Group("/socials").Use(authentication)
	{
		telegramController := &controllers.TelegramController{
			DB:    db,
			Redis: redis,
		}
		//companyIDPathName := "companyID"

		// authentication data
		telegrams.POST("/telegram/bind", telegramController.AuthenticateAction)
	}
}
