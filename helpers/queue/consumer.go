package queue

import (
	"encoding/json"
	"fmt"
	"go-api/configs"
	"log"

	"github.com/streadway/amqp"
)

type ConsumerQueueHandler func(map[string]interface{})

type BaseQueueConsumer struct {
	configConsumerQueueDeclare ConsumerQueueDeclareConfig
	configConsumer             ConsumerConfig

	queueConnection *amqp.Connection
	queueChannel    *amqp.Channel
}

type ConsumerQueueDeclareConfig struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp.Table
}

type ConsumerConfig struct {
	QueueName string
	Consumer  string
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Args      amqp.Table
}

func BaseConsumer() *BaseQueueConsumer {
	connection, channel := configs.GetRabbitQueue()
	baseQueueConsumer := &BaseQueueConsumer{
		configConsumerQueueDeclare: ConsumerQueueDeclareConfig{},
		configConsumer:             ConsumerConfig{},
		queueConnection:            connection,
		queueChannel:               channel,
	}

	return baseQueueConsumer
}

func (base *BaseQueueConsumer) GetDefaultConfigQueueDeclare() ConsumerQueueDeclareConfig {
	return ConsumerQueueDeclareConfig{
		Name:       "test",
		Durable:    false,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       nil,
	}
}

func (base *BaseQueueConsumer) GetDefaultConfigConsumer() ConsumerConfig {
	return ConsumerConfig{
		QueueName: "test",
		Consumer:  "",
		AutoAck:   true,
		Exclusive: false,
		NoLocal:   false,
		NoWait:    false,
		Args:      nil,
	}
}

func (base *BaseQueueConsumer) SetConfigQueue(config ConsumerQueueDeclareConfig) *BaseQueueConsumer {
	base.configConsumerQueueDeclare = config
	return base
}

func (base *BaseQueueConsumer) SetConfigConsumer(config ConsumerConfig) *BaseQueueConsumer {
	base.configConsumer = config
	return base
}

func (base *BaseQueueConsumer) Consume(handler ConsumerQueueHandler) error {
	channel := base.queueChannel
	queueDeclareConfig := base.configConsumerQueueDeclare
	queueConsumerConfig := base.configConsumer

	queueDeclare, err := channel.QueueDeclare(
		queueDeclareConfig.Name,       // queue name
		queueDeclareConfig.Durable,    // durable
		queueDeclareConfig.AutoDelete, // delete queue when unused
		queueDeclareConfig.Exclusive,  // exclusive
		queueDeclareConfig.NoWait,     // no wait
		queueDeclareConfig.Args,       // argument
	)
	if err != nil {
		return err
	}

	messages, err := channel.Consume(
		queueDeclare.Name,
		queueConsumerConfig.Consumer,
		queueConsumerConfig.AutoAck,
		queueConsumerConfig.Exclusive,
		queueConsumerConfig.NoLocal,
		queueConsumerConfig.NoWait,
		queueConsumerConfig.Args,
	)
	if err != nil {
		return err
	}

	forever := make(chan bool)

	go func() {
		for message := range messages {
			log.Printf(fmt.Sprint(message))

			var mapData map[string]interface{}
			_ = json.Unmarshal(message.Body, &mapData)
			handler(mapData)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
	return nil
}
