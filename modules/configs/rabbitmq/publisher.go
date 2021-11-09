package rabbitmq

import (
	"os"
	"time"

	"github.com/streadway/amqp"
)

type PublisherConfig struct {
	Mandatory bool            `json:"mandatory"`
	Immediate bool            `json:"immediate"`
	Msg       amqp.Publishing `json:"msg"`
}

func (base *QueueSetup) AddPublisher(queueDeclare *QueueDeclareConfig, publisherConfig *PublisherConfig) *QueueSetup {
	if queueDeclare == nil { // set default configuration queue declare
		queueDeclare = &QueueDeclareConfig{
			Durable:    false,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Args:       nil,
		}
	}

	if publisherConfig == nil { // set default configuration publisher config
		publisherConfig = &PublisherConfig{
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

	base.queueConfig = &QueueConfig{
		QueueDeclareConfig:   queueDeclare,
		QueuePublisherConfig: publisherConfig,
	}

	err := base.declareQueue()
	if err != nil {
		loggingMessage("error declare queue after open connection", err.Error())
		panic(err.Error())
	}

	go base.reconnect(false)

	return base
}

func (base *QueueSetup) Publish(exchangeName, message string) error {
	publishConfig := base.queueConfig.QueuePublisherConfig
	publishConfig.Msg.Body = []byte(message)

	loggingMessage("Publishing Message...", nil)
	err := base.channel.Publish(
		exchangeName,
		base.queueName,
		publishConfig.Mandatory,
		publishConfig.Immediate,
		publishConfig.Msg,
	)

	return err
}

func (base *QueueSetup) BatchPublish(exchangeName string, messages []string) []error {
	publishConfig := base.queueConfig.QueuePublisherConfig

	var listErr []error
	for _, message := range messages {
		publishConfig.Msg.Body = []byte(message)

		loggingMessage("Publishing Message...", nil)
		err := base.channel.Publish(
			exchangeName,
			base.queueName,
			publishConfig.Mandatory,
			publishConfig.Immediate,
			publishConfig.Msg,
		)

		if err != nil {
			listErr = append(listErr, err)
		}
	}

	return listErr
}
