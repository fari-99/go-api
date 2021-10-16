package rabbitmq

import (
	"github.com/gin-gonic/gin"
	"go-api/modules/configs"
	"log"
)

func NewRoute(app *gin.Engine) {
	log.Println("Setup Test RabbitMq Connection router")

	testRabbit := app.Group("/test-rabbit")
	{
		testRabbitMqQueueController := &RabbitMqController{
			DI: configs.DIInit(),
		}

		testRabbit.POST("/queue", testRabbitMqQueueController.TestPublishQueueAction)
		testRabbit.POST("/batch-queue", testRabbitMqQueueController.TestBatchPublishQueueAction)
	}
}
