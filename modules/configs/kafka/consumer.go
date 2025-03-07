package kafka

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/IBM/sarama"
)

type kafkaConsumerUtil struct {
	consumer sarama.ConsumerGroup
}

var kafkaConsumerInstance *kafkaConsumerUtil
var onceKafkaConsumer sync.Once

// GetKafkaConsumerConnection get kafka connection
func GetKafkaConsumerConnection(configKafka ConfigKafka, groupID string) sarama.ConsumerGroup {
	onceKafkaConsumer.Do(func() {
		consumer, err := sarama.NewConsumerGroup([]string{configKafka.DSN}, groupID, configKafka.Config)
		if err != nil {
			panic(err)
		}

		kafkaConsumerInstance = &kafkaConsumerUtil{
			consumer: consumer,
		}
	})

	return kafkaConsumerInstance.consumer
}

type EventHandler interface {
	Setup(sarama.ConsumerGroupSession) error
	Cleanup(sarama.ConsumerGroupSession) error
	ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error
}

type Consumer interface {
	Subscribe(handler EventHandler)
	Unsubscribe()
}

type kafkaConsumer struct {
	topic         []string
	consumerGroup sarama.ConsumerGroup
}

func NewConsumer(consumerGroupID string, topic []string) (Consumer, error) {
	kafkaConfig := getKafkaConfig(TypeProducerKafka)
	if topic == nil || len(topic) == 0 {
		topic = []string{os.Getenv("KAFKA_DEFAULT_TOPIC")}
	}

	cg := GetKafkaConsumerConnection(kafkaConfig, consumerGroupID)
	return &kafkaConsumer{
		topic:         topic,
		consumerGroup: cg,
	}, nil
}

func (c *kafkaConsumer) Subscribe(handler EventHandler) {
	ctx, cancel := context.WithCancel(context.Background())
	topics := func() []string {
		result := make([]string, 0)
		result = append(result, c.topic...)
		return result
	}

	go func() {
		for {
			if err := c.consumerGroup.Consume(ctx, topics(), handler); err != nil {
				log.Fatal("Error from consumer : ", err.Error())
			}

			if ctx.Err() != nil {
				log.Fatal("Error from consumer : ", ctx.Err().Error())
			}
		}
	}()

	go func() {
		for err := range c.consumerGroup.Errors() {
			log.Println("Error from consumer group : ", err.Error())
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("terminating: context cancelled")
				cancel()
			}
		}
	}()
	log.Printf("Kafka consumer listens topic : %v \n", c.topic)
}

func (c *kafkaConsumer) Unsubscribe() {
	if err := c.consumerGroup.Close(); err != nil {
		log.Printf("Client wasn't closed :%+v", err)
	}
	log.Println("Kafka consumer closed")
}
