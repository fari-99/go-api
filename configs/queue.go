package configs

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

type RabbitQueue struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

const delayReconnectQueue = 2

var rabbitQueueInstance *RabbitQueue
var rabbitQueueOnce sync.Once

func GetRabbitQueue() (*amqp.Connection, *amqp.Channel) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic setup RabbitMQ Queue, ", r)
		}
	}()

	rabbitQueueOnce.Do(func() {
		log.Println("Initialize RabbitMq Queue connection...")

		connection, err := GetRabbitMqQueueConnection()
		if err != nil {
			panic(err.Error())
		}

		channel, err := connection.Channel()
		if err != nil {
			panic(err.Error())
		}

		rabbitQueueInstance = &RabbitQueue{
			Connection: connection,
			Channel:    channel,
		}
	})

	return rabbitQueueInstance.Connection, rabbitQueueInstance.Channel
}

func GetRabbitMqQueueConnection() (*amqp.Connection, error) {
	connUrl, err := getRabbitConfig()
	if err != nil {
		return &amqp.Connection{}, err
	}

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

	return connection, err
}

// RabbitHost
type RabbitHost struct {
	RabbitHost, RabbitPort, RabbitUser, RabbitPassword string
}

func getRabbitConfig() (connURL string, err error) {
	var idx int
	con := make(map[int]RabbitHost)
	con[0] = RabbitHost{
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

		con[idx] = RabbitHost{
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
