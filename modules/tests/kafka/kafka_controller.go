package kafka

import (
	"fmt"
	"go-api/helpers"
	"go-api/modules/configs"
	"go-api/modules/configs/kafka"
	"log"
	"net/http"
	"os"

	"github.com/IBM/sarama"

	"github.com/gin-gonic/gin"
)

type KafkaController struct {
	*configs.DI
}

func (controller *KafkaController) KafkaPublishSingleTest(ctx *gin.Context) {
	kafkaProducer := kafka.GetKafkaProducerConnection()
	kafkaMsg := &sarama.ProducerMessage{
		Topic: os.Getenv("KAFKA_DEFAULT_TOPIC"),
		Value: sarama.StringEncoder("im here"),
	}

	partition, offset, err := kafkaProducer.SendMessage(kafkaMsg)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("Topic %v, Partition %v, Offset %d", kafkaMsg.Topic, partition, offset)
	helpers.NewResponse(ctx, http.StatusOK, "success publish message to kafka")
	return
}

func (controller *KafkaController) KafkaPublishBatchTest(ctx *gin.Context) {
	kafkaProducer := kafka.GetKafkaProducerConnection()
	var messages []*sarama.ProducerMessage
	for i := 0; i < 10; i++ {
		kafkaMsg := &sarama.ProducerMessage{
			Topic: os.Getenv("KAFKA_DEFAULT_TOPIC"),
			Value: sarama.StringEncoder(fmt.Sprintf("im here %d", i)),
		}

		messages = append(messages, kafkaMsg)
	}

	err := kafkaProducer.SendMessages(messages)
	if err != nil {
		helpers.NewResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	helpers.NewResponse(ctx, http.StatusOK, "success publish message to kafka")
	return
}
