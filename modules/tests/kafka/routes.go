package kafka

import (
	"log"

	"github.com/gin-gonic/gin"
)

func NewRoute(app *gin.Engine) {
	log.Println("Setup Test Kafka router")

	testCrypt := app.Group("/test-kafka")
	{
		kafkaController := &KafkaController{}

		testCrypt.POST("/publish/single", kafkaController.KafkaPublishSingleTest)
		testCrypt.POST("/publish/batch", kafkaController.KafkaPublishBatchTest)
	}
}
