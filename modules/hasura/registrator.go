package hasura

import (
	"log"

	"github.com/gin-gonic/gin"
)

func NewRegistrator(app *gin.RouterGroup, service Service) {
	log.Println("Setup Hasura router")
	control := controller{service: service}

	events := app.Group("/hasura/events")
	{
		events.PUT("/articles", control.UpdateArticle)
	}

	cron := app.Group("/hasura/crons")
	{
		cron.POST("/articles", control.UpdateArticle)
	}

	schedule := app.Group("/hasura/schedule")
	{
		schedule.POST("/articles", control.UpdateArticle)
	}
}
