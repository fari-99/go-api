package configs

import (
	"errors"
	"fmt"
	"go-api/constant"
	"go-api/helpers"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/kataras/iris"

	"github.com/streadway/amqp"
)

// TODO: split queue connection and channel, so it can be reused

type QueueSetup struct {
	queueName  string
	connection *amqp.Connection
	channel    *amqp.Channel
	closed     bool

	errorConnection chan *amqp.Error

	queueConfig   *QueueConfig
	queueConsumer ConsumerHandler
}

type ConsumerHandler func(string)

type QueueConfig struct {
	QueueDeclareConfig   *QueueDeclareConfig
	QueueConsumerConfig  *ConsumerConfig
	QueuePublisherConfig *PublisherConfig
}

type QueueDeclareConfig struct {
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp.Table
}

type ConsumerConfig struct {
	Consumer  string
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Args      amqp.Table
}

type PublisherConfig struct {
	Mandatory bool            `json:"mandatory"`
	Immediate bool            `json:"immediate"`
	Msg       amqp.Publishing `json:"msg"`
}

type RabbitHost struct {
	RabbitHost     string
	RabbitPort     string
	RabbitUser     string
	RabbitPassword string
}

func NewBaseQueue() *QueueSetup {
	queueSetup := QueueSetup{
		queueName:       constant.QueueDefaultName,
		connection:      nil,
		channel:         nil,
		closed:          false,
		errorConnection: nil,
		queueConfig:     nil,
	}

	return &queueSetup
}

func (base *QueueSetup) SetQueueName(queueName string) *QueueSetup {
	if len(queueName) == 0 || queueName == "" {
		queueName = constant.QueueDefaultName
	}

	base.queueName = queueName
	return base
}

func (base *QueueSetup) AddConsumer(queueDeclare *QueueDeclareConfig, consumerConfig *ConsumerConfig) *QueueSetup {
	if queueDeclare == nil { // set default configuration queue declare
		queueDeclare = &QueueDeclareConfig{
			Durable:    false,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Args:       nil,
		}
	}

	if consumerConfig == nil {
		consumerConfig = &ConsumerConfig{
			Consumer:  "",
			AutoAck:   true,
			Exclusive: false,
			NoLocal:   false,
			NoWait:    false,
			Args:      nil,
		}
	}

	base.queueConfig = &QueueConfig{
		QueueDeclareConfig:  queueDeclare,
		QueueConsumerConfig: consumerConfig,
	}

	err := base.openConnection()
	if err != nil {
		helpers.LoggingMessage("error open connection", err.Error())
		panic(err.Error())
	}

	go base.reconnect()

	return base
}

func (base *QueueSetup) Consume(consumer ConsumerHandler) {
	helpers.LoggingMessage("Registering Consumer...", nil)
	deliveries, err := base.registerQueueConsumer()
	if err != nil {
		helpers.LoggingMessage("Error register queue queueConsumer", err.Error())
		panic(err.Error())
	}

	base.executeMessageConsumer(consumer, deliveries, false)
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

	err := base.openConnection()
	if err != nil {
		helpers.LoggingMessage("error open connection", err.Error())
		panic(err.Error())
	}

	go base.reconnect()

	return base
}

func (base *QueueSetup) Publish(message string) error {
	publishConfig := base.queueConfig.QueuePublisherConfig
	publishConfig.Msg.Body = []byte(message)

	helpers.LoggingMessage("Pubilshing Message...", nil)
	err := base.channel.Publish(
		"",
		base.queueName,
		publishConfig.Mandatory,
		publishConfig.Immediate,
		publishConfig.Msg,
	)

	return err
}

func (base *QueueSetup) Close() {
	log.Println("Closing connection")
	base.closed = true
	_ = base.channel.Close()
	_ = base.connection.Close()
}

func (base *QueueSetup) reconnect() {
	tryReconnect := 1
	for {
		helpers.LoggingMessage("Try Reconnecting", tryReconnect)

		err := <-base.errorConnection
		if !base.closed || err != nil {
			if err != nil {
				helpers.LoggingMessage("Reconnecting after connection closed", err.Error())
			}

			errorData := base.openConnection()
			if errorData != nil {
				helpers.LoggingMessage("-Failed- Open Connection after connection closed", err.Error())
				continue
			}

			go base.reconnect()

			errorData = base.recoverQueueConsumers()
			if errorData != nil {
				helpers.LoggingMessage("-Failed- Recover Queue after connection closed", err.Error())
				continue
			}

			break
		}

		tryReconnect++
	}
}

func (base *QueueSetup) recoverQueueConsumers() error {
	var consumer = base.queueConsumer

	log.Println("Recovering queueConsumer...")
	messages, err := base.registerQueueConsumer()
	if err != nil {
		return err
	}

	helpers.LoggingMessage("Consumer recovered! Continuing message processing...", nil)
	base.executeMessageConsumer(consumer, messages, true)
	return nil
}

func (base *QueueSetup) executeMessageConsumer(consumer ConsumerHandler, deliveries <-chan amqp.Delivery, isRecovery bool) {
	helpers.LoggingMessage("Consumer successfully registered, processing messages...", nil)

	if !isRecovery {
		base.queueConsumer = consumer
	}

	forever := make(chan bool)

	go func() {
		helpers.LoggingMessage("Consumer Ready", iris.Map{"PID": os.Getpid()})

		isAutoAck := base.queueConfig.QueueConsumerConfig.AutoAck

		for delivery := range deliveries {
			consumer(string(delivery.Body[:]))

			if !isAutoAck {
				if err := delivery.Ack(false); err != nil {
					log.Printf("Error acknowledging message : %s", err)
				} else {
					log.Printf("Acknowledged message")
				}
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
	return
}

func (base *QueueSetup) openConnection() error {
	tryReconnect := 1
	totalTry, _ := strconv.ParseInt(os.Getenv("RABBIT_RECONNECT"), 10, 64)

	for {
		if int(totalTry) == tryReconnect {
			break
		}

		connUrl, err := base.getRabbitConfig()
		if err != nil {
			tryReconnect++
			helpers.LoggingMessage("Error get config connection to RabbitMq", err.Error())
			continue
		}

		helpers.LoggingMessage("Connecting to RabbitMq", connUrl)
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
			tryReconnect++
			helpers.LoggingMessage("Error get config connection to RabbitMq", err.Error())
			continue
		}

		base.connection = connection
		base.errorConnection = make(chan *amqp.Error)
		base.connection.NotifyClose(base.errorConnection)

		err = base.openChannel()
		if err != nil {
			tryReconnect++
			helpers.LoggingMessage("Error open channel", err.Error())
			continue
		}

		if base.queueConfig == nil {
			panic("Configuration queueConsumer is empty")
		}

		err = base.declareQueue()
		if err != nil {
			tryReconnect++
			helpers.LoggingMessage("Error declare queue", err.Error())
			continue
		}

		helpers.LoggingMessage("Connection RabbitMq Established!!", nil)
		break
	}

	if tryReconnect == int(totalTry) {
		err := fmt.Errorf("reconnect exceed max reconnect value")
		return err
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

// RabbitHost
type RabbitHosts struct {
	RabbitHost, RabbitPort, RabbitUser, RabbitPassword string
}

func (base *QueueSetup) getRabbitConfig() (connURL string, err error) {
	var idx int
	con := make(map[int]RabbitHosts)
	con[0] = RabbitHosts{
		RabbitHost:     os.Getenv("RABBIT_HOST"),
		RabbitPort:     os.Getenv("RABBIT_PORT"),
		RabbitUser:     os.Getenv("RABBIT_USER"),
		RabbitPassword: os.Getenv("RABBIT_PASSWORD"),
	}

	if len(con[0].RabbitHost) == 0 {
		err = errors.New("environment variable for RABBIT_HOST is not found")
	} else if len(con[0].RabbitUser) == 0 {
		err = errors.New("environment variable for RABBIT_USER is not found")
	}

	idx = 1
	for {

		value := os.Getenv("RABBIT_HOST" + strconv.Itoa(idx))
		if len(value) == 0 {
			break
		}

		con[idx] = RabbitHosts{
			RabbitHost:     os.Getenv("RABBIT_HOST" + strconv.Itoa(idx)),
			RabbitPort:     os.Getenv("RABBIT_PORT" + strconv.Itoa(idx)),
			RabbitUser:     os.Getenv("RABBIT_USER" + strconv.Itoa(idx)),
			RabbitPassword: os.Getenv("RABBIT_PASSWORD" + strconv.Itoa(idx)),
		}

		idx++
	}

	rand.Seed(time.Now().UnixNano())
	random := rand.Intn(len(con))
	connURL = fmt.Sprintf("amqp://%s:%s@%s:%s/",
		con[random].RabbitUser,
		con[random].RabbitPassword,
		con[random].RabbitHost,
		con[random].RabbitPort)

	return connURL, err
}
