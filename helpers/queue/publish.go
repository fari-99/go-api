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
		Name:       "test",
		Durable:    false,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       nil,
	}
}

func (base *BaseQueue) GetDefaultConfigPublishQueue() ConfigPublishQueue {
	return ConfigPublishQueue{
		Key:       "test",
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
		return errors.New("confirm.select destination: " + err.Error())
	}

	go func(connection *amqp.Connection, channel *amqp.Channel) {
		_ = base.sendMessages(connection, channel, message)
		confirmed := <-confirms
		if confirmed.Ack {
			log.Printf("confirmed delivery with delivery tag: %d", confirmed.DeliveryTag)
		} else {
			log.Printf("failed delivery of delivery tag: %d", confirmed.DeliveryTag)
			for {
				err := base.TryReconnect(message)
				if err == nil {
					break
				}
			}
		}
	}(connection, channel)

	return nil
}

func (base *BaseQueue) BatchPublish(messages []string) []string {
	var allError []string

	connection := base.queueConnection
	channel := base.queueChannel

	confirms := channel.NotifyPublish(make(chan amqp.Confirmation, 1))
	if err := channel.Confirm(false); err != nil {
		allError = append(allError, "confirm.select destination: "+err.Error())
		return allError
	}

	testConnection := 0
	for _, message := range messages {
		if !helpers.IsJSON(message) {
			allError = append(allError, "data is not json string, err := "+message)
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
				for {
					err := base.TryReconnect(message)
					if err == nil {
						break
					}
				}
			}
		}(connection, channel, message)

		if testConnection == 5 {
			_ = connection.Close()
		} else {
			testConnection++
		}
	}

	return allError
}

func (base *BaseQueue) TryReconnect(message string) error {
	time.Sleep(5 * time.Second)

	connection, err := configs.GetRabbitMqQueueConnection()
	if err != nil {
		return err
	}

	channel, err := connection.Channel()
	if err != nil {
		return err
	}

	base.queueConnection = connection
	base.queueChannel = channel

	err = base.PublishQueue(message)
	return err
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
