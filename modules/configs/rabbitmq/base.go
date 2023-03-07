package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueSetup struct {
	exchangeName string
	queueName    string
	connection   *amqp.Connection
	channel      *amqp.Channel
	closed       bool

	errorConnection chan *amqp.Error

	queueConfig    *QueueConfig
	queueConsumer  ConsumerHandler
	exchangeConfig *ExchangeConfig

	ctx context.Context
}

type QueueConfig struct {
	QueueDeclareConfig   *QueueDeclareConfig
	QueueConsumerConfig  *ConsumerConfig
	QueuePublisherConfig *PublisherConfig
	QueueBindConfig      *QueueBindConfig
}

type ExchangeConfig struct {
	ExchangeDeclareConfig *ExchangeDeclareConfig
}

type QueueDeclareConfig struct {
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp.Table
}

type QueueBindConfig struct {
	RoutingKey string
	NoWait     bool
	Args       amqp.Table
}

func NewBaseQueue(exchangeName, queueName string) *QueueSetup {
	ctx := context.TODO()

	queueSetup := &QueueSetup{
		exchangeName: exchangeName,
		ctx:          ctx, // default ctx
	}

	queueSetup.setQueueName(queueName)
	newQueueSetup := queueSetup.setQueueUtil()

	return newQueueSetup
}

func (base *QueueSetup) SetContext(ctx context.Context) *QueueSetup {
	base.ctx = ctx
	return base
}

func (base *QueueSetup) setQueueUtil() *QueueSetup {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic setup queue, ", r)
		}
	}()

	err := base.openConnection()
	if err != nil {
		loggingMessage("error open connection", err.Error())
		panic(err.Error())
	}

	return base
}

func (base *QueueSetup) setQueueName(queueName string) *QueueSetup {
	base.queueName = queueName
	if len(base.queueName) == 0 || base.queueName == "" {
		base.queueName = os.Getenv("DEFAULT_QUEUE_NAME")
	}

	return base
}

func (base *QueueSetup) setExchangeName() *QueueSetup {
	if len(base.exchangeName) == 0 || base.exchangeName == "" {
		base.exchangeName = os.Getenv("DEFAULT_EXCHANGE")
	}

	return base
}

func (base *QueueSetup) declareQueue() error {
	queueDeclareConfig := base.queueConfig.QueueDeclareConfig

	_, err := base.channel.QueueDeclare(
		base.queueName,
		queueDeclareConfig.Durable,
		queueDeclareConfig.AutoDelete,
		queueDeclareConfig.Exclusive,
		queueDeclareConfig.NoWait,
		queueDeclareConfig.Args,
	)

	if err != nil {
		return err
	}

	return nil
}

func (base *QueueSetup) Close() {
	loggingMessage("Closing Connection", nil)
	base.closed = true
	err := base.channel.Close()
	if err != nil {
		loggingMessage("Error closing channel", err.Error())
	}

	err = base.connection.Close()
	if err != nil {
		loggingMessage("Error closing connection", err.Error())
	}
}

func (base *QueueSetup) reconnect(isRecover bool) {
	for {
		loggingMessage("Trying to reconnect, please wait...", nil)
		time.Sleep(3 * time.Second)

		err := <-base.errorConnection
		if !base.closed || err != nil {
			if err != nil {
				loggingMessage("Reconnecting after connection closed", err.Error())
			}

			if base.exchangeName != "" {
				base.AddConsumerExchange(true)
			} else {
				base.AddConsumer(true)
			}

			go base.reconnect(isRecover)

			if isRecover {
				errorData := base.recoverQueueConsumers()
				if errorData != nil {
					loggingMessage("-Failed- Recover Queue after connection closed", errorData.Error())
					continue
				}
			}

			break
		}
	}

	loggingMessage("Success reconnect...\n\n", nil)
}

func (base *QueueSetup) recoverQueueConsumers() error {
	var consumer = base.queueConsumer

	loggingMessage("Recovering queueConsumer...", nil)
	messages, err := base.registerQueueConsumer()
	if err != nil {
		return err
	}

	loggingMessage("Consumer recovered! Continuing message processing...", nil)
	base.executeMessageConsumer(consumer, messages, true)
	return nil
}

func (base *QueueSetup) registerQueueConsumer() (<-chan amqp.Delivery, error) {
	consumerConfig := base.queueConfig.QueueConsumerConfig
	message, err := base.channel.Consume(
		base.queueName,
		consumerConfig.Consumer,
		consumerConfig.AutoAck,
		consumerConfig.Exclusive,
		consumerConfig.NoLocal,
		consumerConfig.NoWait,
		consumerConfig.Args,
	)
	return message, err
}

func (base *QueueSetup) executeMessageConsumer(consumer ConsumerHandler, deliveries <-chan amqp.Delivery, isRecovery bool) {
	if !isRecovery {
		base.queueConsumer = consumer
	}

	forever := make(chan bool)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				loggingMessage("Recovered from panic on your queue", r)
				base.Close()
				_ = base.openConnection()
				_ = base.recoverQueueConsumers()
			}
		}()

		loggingMessage("Consumer Ready", gin.H{"PID": os.Getpid()})

		isAutoAck := base.queueConfig.QueueConsumerConfig.AutoAck

		for delivery := range deliveries {
			var handlerData ConsumerHandlerData
			_ = json.Unmarshal(delivery.Body[:], &handlerData)

			consumer(handlerData)
			if !isAutoAck {
				if err := delivery.Ack(false); err != nil {
					loggingMessage("Error acknowledging message", err.Error())
				} else {
					loggingMessage("Acknowledged message", nil)
				}
			}
		}
	}()

	loggingMessage(" [*] Waiting for messages. To exit press CTRL+C", nil)
	<-forever
	return
}

func (base *QueueSetup) openConnection() error {
	for {
		loggingMessage("Trying to open rabbitmq connection, please wait...", nil)
		time.Sleep(5 * time.Second)

		connUrl := fmt.Sprintf("amqp://%s:%s@%s:%s/",
			os.Getenv("RABBIT_USER"),
			os.Getenv("RABBIT_PASSWORD"),
			os.Getenv("RABBIT_HOST"),
			os.Getenv("RABBIT_PORT"))

		connection, err := amqp.DialConfig(connUrl, amqp.Config{
			//SASL:            nil,
			//Vhost:           "",
			//ChannelMax:      0,
			//FrameSize:       0,
			Heartbeat: 10 * time.Second, // default value
			//TLSClientConfig: nil,
			//Properties:      nil,
			//Locale:          "en_US",
			//Dial:            nil,
		})

		if err != nil {
			loggingMessage("Error get config connection to RabbitMq", err.Error())
			continue
		}

		base.connection = connection
		base.errorConnection = make(chan *amqp.Error)
		base.connection.NotifyClose(base.errorConnection)

		err = base.openChannel()
		if err != nil {
			loggingMessage("Error open channel", err.Error())
			continue
		}

		loggingMessage("Connection RabbitMq Established!!", nil)
		break
	}

	return nil
}

func (base *QueueSetup) openChannel() error {
	channel, err := base.connection.Channel()
	if err != nil {
		return err
	}

	base.channel = channel
	return nil
}

func loggingMessage(message string, data interface{}) {
	if data != nil {
		dataMarshal, _ := json.Marshal(data)
		message += fmt.Sprintf("%s, Data := %s", message, string(dataMarshal))
	}

	if os.Getenv("APP_MODE") == gin.DebugMode {
		log.Printf(message)
	}
}
