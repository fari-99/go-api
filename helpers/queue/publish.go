package queue

import (
	"errors"
	"go-api/configs"
	"go-api/helpers"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
)

type BaseQueue struct {
	configQueue        ConfigQueue
	configPublishQueue ConfigPublishQueue

	queueConnection *amqp.Connection
	queueChannel    *amqp.Channel
}

type ConfigQueue struct {
	Name       string     `json:"name"`
	Durable    bool       `json:"durable"`
	AutoDelete bool       `json:"auto_delete"`
	Exclusive  bool       `json:"exclusive"`
	NoWait     bool       `json:"no_wait"`
	Args       amqp.Table `json:"args"`
}

type ConfigPublishQueue struct {
	Key       string          `json:"key"`
	Mandatory bool            `json:"mandatory"`
	Immediate bool            `json:"immediate"`
	Msg       amqp.Publishing `json:"msg"`
}

func NewBaseQueue() (*BaseQueue, error) {
	connection, channel := configs.GetRabbitQueue()
	baseQueue := &BaseQueue{
		configQueue:        ConfigQueue{},
		configPublishQueue: ConfigPublishQueue{},
		queueConnection:    connection,
		queueChannel:       channel,
	}

	return baseQueue, nil
}

func (base *BaseQueue) GetDefaultConfigQueue() ConfigQueue {
	return ConfigQueue{
		Name:       "your-queue-name-here",
		Durable:    false,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       nil,
	}
}

func (base *BaseQueue) GetDefaultConfigPublishQueue() ConfigPublishQueue {
	return ConfigPublishQueue{
		Key:       "your-queue-name-here",
		Mandatory: false,
		Immediate: false,
		Msg: amqp.Publishing{
			//Headers:         nil,
			ContentType: "application/json",
			//ContentEncoding: "",
			DeliveryMode: 1,
			//Priority:        0,
			//CorrelationId:   "",
			//ReplyTo:         "",
			//Expiration:      "",
			//MessageId:       "",
			Timestamp: time.Now(),
			//Type:            "",
			//UserId:          "",
			AppId: os.Getenv("APP_NAME"),
			//Body:            nil,
		},
	}
}

func (base *BaseQueue) SetConfigQueue(config ConfigQueue) *BaseQueue {
	base.configQueue = config
	return base
}

func (base *BaseQueue) SetConfigPublishQueue(config ConfigPublishQueue) *BaseQueue {
	base.configPublishQueue = config
	return base
}

func (base *BaseQueue) PublishQueue(message string) error {
	if !helpers.IsJSON(message) {
		return errors.New("data is not json string")
	}

	connection := base.queueConnection
	channel := base.queueChannel

	confirms := channel.NotifyPublish(make(chan amqp.Confirmation, 1))
	if err := channel.Confirm(false); err != nil {
		log.Fatalf("confirm.select destination: %s", err)
	}

	go func(connection *amqp.Connection, channel *amqp.Channel) {
		_ = base.sendMessages(connection, channel, message)
		confirmed := <-confirms
		if confirmed.Ack {
			log.Printf("confirmed delivery with delivery tag: %d", confirmed.DeliveryTag)
		} else {
			log.Printf("failed delivery of delivery tag: %d", confirmed.DeliveryTag)
			// TODO: Try reconnect here
		}
	}(connection, channel)

	return nil
}

func (base *BaseQueue) BatchPublish(messages []string) []error {
	var allError []error

	connection := base.queueConnection
	channel := base.queueChannel

	confirms := channel.NotifyPublish(make(chan amqp.Confirmation, 1))
	if err := channel.Confirm(false); err != nil {
		log.Fatalf("confirm.select destination: %s", err)
	}

	for _, message := range messages {
		if !helpers.IsJSON(message) {
			allError = append(allError, errors.New("data is not json string, err := "+message))
			continue
		}

		go func(connection *amqp.Connection, channel *amqp.Channel, message string) {
			_ = base.sendMessages(connection, channel, message)

			// only ack the source delivery when the destination acks the publishing
			confirmed := <-confirms
			if confirmed.Ack {
				log.Printf("confirmed delivery with delivery tag: %d", confirmed.DeliveryTag)
			} else {
				log.Printf("failed delivery of delivery tag: %d", confirmed.DeliveryTag)
				// TODO: Try reconnect here
			}
		}(connection, channel, message)
	}

	return allError
}

func (base *BaseQueue) sendMessages(connection *amqp.Connection, channel *amqp.Channel, message string) error {
	queueConfig := base.configQueue

	_, err := channel.QueueDeclare(
		queueConfig.Name,
		queueConfig.Durable,
		queueConfig.AutoDelete,
		queueConfig.Exclusive,
		queueConfig.NoWait,
		queueConfig.Args,
	)

	if err != nil {
		return err
	}

	queuePublishConfig := base.configPublishQueue
	queuePublishConfig.Msg.Body = []byte(message)
	err = channel.Publish(
		"",
		queuePublishConfig.Key,
		queuePublishConfig.Mandatory,
		queuePublishConfig.Immediate,
		queuePublishConfig.Msg,
	)

	return err
}
