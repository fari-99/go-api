package kafka

import (
	"sync"

	"github.com/IBM/sarama"
)

type kafkaProducerUtil struct {
	producer sarama.SyncProducer
}

var kafkaProducerInstance *kafkaProducerUtil

var onceKafkaProducer sync.Once

// GetKafkaProducerConnection get kafka producer connection
func GetKafkaProducerConnection() sarama.SyncProducer {
	kafkaConfig := getKafkaConfig(TypeProducerKafka)

	onceKafkaProducer.Do(func() {
		kafkaSync, err := sarama.NewSyncProducer([]string{kafkaConfig.DSN}, kafkaConfig.Config)
		if err != nil {
			panic(err)
		}

		kafkaProducerInstance = &kafkaProducerUtil{
			producer: kafkaSync,
		}
	})

	return kafkaProducerInstance.producer
}
