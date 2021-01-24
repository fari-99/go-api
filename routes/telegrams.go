package routes

import (
	"go-api/controllers"
	"go-api/middleware"
	"log"

	"github.com/kataras/iris/v12"
)

func (routes *Routes) setupTelegramRoute() *iris.Application {
	log.Println("Setup Telegram router")

	app := routes.irisApp
	db := routes.DB
	redis := routes.Redis

	authentication := middleware.NewMiddleware(middleware.BaseMiddleware{})

	// Approver Endpoint collection
	app.PartyFunc("/telegrams", func(telegrams iris.Party) {
		telegramController := &controllers.TelegramController{
			DB:    db,
			Redis: redis,
		}
		//companyIDPathName := "companyID"

		// authentication data
		telegrams.Post("/authenticate", authentication, telegramController.AuthenticateAction)
	})

	return app
}
