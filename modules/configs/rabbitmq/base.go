package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
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

	isPublisher         bool
	maxReconnectAttempt int
	reconnectAttempt    int

	ctx    context.Context
	cancel context.CancelFunc

	waitGroup sync.WaitGroup
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
	Durable    bool       // If true, the queue will survive broker restarts
	AutoDelete bool       // If true, the queue is deleted when last consumer unsubscribes
	Exclusive  bool       // If true, the queue can only be used by the declaring connection
	NoWait     bool       // If true, the server won't respond to the method (fire-and-forget)
	Args       amqp.Table // Optional arguments (e.g., message TTL, DLX)
}

type QueueBindConfig struct {
	RoutingKey string
	NoWait     bool
	Args       amqp.Table
}

func NewBaseQueue(exchangeName, queueName string) *QueueSetup {
	log.Println("Initialize RabbitMQ Queue connection...")
	ctx, cancel := context.WithCancel(context.Background())

	queueSetup := &QueueSetup{
		exchangeName: exchangeName,
		ctx:          ctx, // default ctx
		cancel:       cancel,
	}

	queueSetup.setQueueName(queueName)
	log.Println("Success Initialize RabbitMQ Queue connection...")
	return queueSetup.setQueueUtil()
}

func (base *QueueSetup) SetContext(ctx context.Context) *QueueSetup {
	newCtx, newCancel := context.WithCancel(ctx)
	base.ctx = newCtx
	base.cancel = newCancel
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
	base.cancel()

	loggingMessage("waiting for consumer done with their process", nil)
	base.waitGroup.Wait()

	if base.channel != nil {
		err := base.channel.Close()
		if err != nil {
			loggingMessage("Error closing channel", err.Error())
		}
	}

	if base.connection != nil {
		err := base.connection.Close()
		if err != nil {
			loggingMessage("Error closing connection", err.Error())
		}
	}

}

func (base *QueueSetup) reconnect() {
	for {
		select {
		case <-base.ctx.Done():
			loggingMessage("Reconnect cancelled", nil)
			return
		case err := <-base.errorConnection:
			if base.closed {
				loggingMessage("Reconnect skipped: connection already closed", nil)
				return
			}

			base.reconnectAttempt++
			if base.isPublisher && base.reconnectAttempt > base.maxReconnectAttempt {
				loggingMessage("Publisher exceeded max reconnect attempts", nil)
				return
			}

			loggingMessage("Reconnecting due to error", err)
			_ = base.openConnection()
			if base.exchangeName != "" {
				base.AddConsumerExchange(true)
			} else {
				base.AddConsumer(true)
			}

			_ = base.recoverQueueConsumers()
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

	base.waitGroup.Add(1)
	go func() {
		defer base.waitGroup.Done()

		defer func() {
			if r := recover(); r != nil {
				loggingMessage("Recovered from panic on your queue", r)
				base.Close()
			}
		}()

		loggingMessage("Consumer Ready", gin.H{"PID": os.Getpid()})

		isAutoAck := base.queueConfig.QueueConsumerConfig.AutoAck

		for {
			select {
			case <-base.ctx.Done():
				loggingMessage("Consumer shutdown via context", nil)
				return
			case delivery, ok := <-deliveries:
				if !ok {
					loggingMessage("Deliveries channel closed", nil)
					return
				}

				var handlerData ConsumerHandlerData
				_ = json.Unmarshal(delivery.Body[:], &handlerData)

				handled := true
				func() {
					defer func() {
						if r := recover(); r != nil {
							loggingMessage("Recovered from panic during message handling", r)
							handled = false
						}
					}()
					consumer(handlerData)
				}()

				if !isAutoAck && handled {
					if err := delivery.Ack(false); err != nil {
						loggingMessage("Error acknowledging message", err.Error())
					} else {
						loggingMessage("Acknowledged message", nil)
					}
				}
			}
		}
	}()

	loggingMessage(" [*] Waiting for messages. To exit press CTRL+C", nil)
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
			// SASL:            nil,
			// Vhost:           "",
			// ChannelMax:      0,
			// FrameSize:       0,
			Heartbeat: 10 * time.Second, // default value
			// TLSClientConfig: nil,
			// Properties:      nil,
			// Locale:          "en_US",
			// Dial:            nil,
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

func (base *QueueSetup) WaitForSignalAndShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sigChan
	loggingMessage("Received shutdown signal", nil)
	base.Close()
}

func loggingMessage(message string, data interface{}) {
	if data != nil {
		dataMarshal, _ := json.Marshal(data)
		message += fmt.Sprintf("%s, Data := %s", message, string(dataMarshal))
	}

	log.Printf(message)
}
