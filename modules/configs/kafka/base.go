package kafka

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Shopify/sarama"
	"github.com/spf13/cast"
)

const TypeConsumerKafka = "consumer"
const TypeProducerKafka = "producer"

type ConfigKafka struct {
	Config *sarama.Config
	DSN    string
}

func getKafkaConfig(kafkaType string) (configKafka ConfigKafka) {
	dsn := fmt.Sprintf("%s:%s", os.Getenv("KAFKA_HOST"), os.Getenv("KAFKA_PORT"))

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Version, _ = sarama.ParseKafkaVersion(os.Getenv("KAFKA_VERSION"))
	kafkaConfig.Net.WriteTimeout = 5 * time.Second

	if cast.ToBool(os.Getenv("KAFKA_SASL_ENABLED")) {
		kafkaConfig.Net.SASL.Enable = true
		kafkaConfig.Net.SASL.User = os.Getenv("KAFKA_USERNAME")
		kafkaConfig.Net.SASL.Password = os.Getenv("KAFKA_PASSWORD")
	}

	switch kafkaType {
	case TypeProducerKafka:
		kafkaConfig.Producer.Retry.Max = 0
		kafkaConfig.Producer.Return.Successes = true
	case TypeConsumerKafka:
		kafkaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
		kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	}

	configKafka = ConfigKafka{
		Config: kafkaConfig,
		DSN:    dsn,
	}

	return configKafka
}

type ConsumerHandlerData struct {
	Date                 time.Time
	Data                 chan interface{}
	SubscriptionStatusCh chan bool
}

func NewConsumerHandler(data chan interface{}) EventHandler {
	return &ConsumerHandlerData{
		Data:                 data,
		SubscriptionStatusCh: make(chan bool),
		Date:                 time.Now(),
	}
}

func (chd *ConsumerHandlerData) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (chd *ConsumerHandlerData) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (chd *ConsumerHandlerData) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		log.Printf("%s", string(message.Value))
		session.MarkMessage(message, "")
	}

	return nil
}
