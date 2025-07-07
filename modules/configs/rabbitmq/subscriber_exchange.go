package rabbitmq

import (
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ExchangeDeclareConfig struct {
	Kind       string
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       amqp.Table
}

/*
SetupExchange this function to help set config exchange binding

	Exchange Default Example
	exchangeDeclareConfig = &ExchangeDeclareConfig{
		Kind:       amqp.ExchangeFanout,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
		Args:       nil,
	}

	Exchange Headers Example
	exchangeDeclareConfig = &ExchangeDeclareConfig{
		Kind:       amqp.ExchangeHeaders,
		Durable:    true,
		AutoDelete: false,
		Internal:   false,
		NoWait:     false,
		Args: amqp.Table{
			"x-match":    "any",
			"event_type": "true",
		},
	}
*/
func (base *QueueSetup) SetupExchange(exchangeDeclareConfig *ExchangeDeclareConfig) *QueueSetup {
	if base.exchangeName == "" {
		base.setExchangeName()
	}

	if exchangeDeclareConfig == nil {
		exchangeDeclareConfig = &ExchangeDeclareConfig{
			Kind:       os.Getenv("DEFAULT_EXCHANGE_TYPE"),
			Durable:    true,
			AutoDelete: false,
			Internal:   false,
			NoWait:     false,
			Args:       nil,
		}
	}

	base.exchangeConfig = &ExchangeConfig{
		ExchangeDeclareConfig: exchangeDeclareConfig,
	}

	return base
}

func (base *QueueSetup) AddConsumerExchange(isReconnect bool) *QueueSetup {
	err := base.openConnection()
	if err != nil {
		loggingMessage("error open new connection", err.Error())
		panic(err.Error())
	}

	err = base.exchangeDeclare()
	if err != nil {
		loggingMessage("error declare exchange after open connection", err.Error())
		panic(err.Error())
	}

	err = base.declareQueue()
	if err != nil {
		loggingMessage("error declare queue after exchange declare", err.Error())
		panic(err.Error())
	}

	err = base.bindQueue()
	if err != nil {
		loggingMessage("error bind queue after queue declare", err.Error())
		panic(err.Error())
	}

	if !isReconnect {
		go base.reconnect()
	}

	return base
}

func (base *QueueSetup) exchangeDeclare() error {
	exchangeDeclareConfig := base.exchangeConfig.ExchangeDeclareConfig

	return base.channel.ExchangeDeclare(
		base.exchangeName,
		exchangeDeclareConfig.Kind,
		exchangeDeclareConfig.Durable,
		exchangeDeclareConfig.AutoDelete,
		exchangeDeclareConfig.Internal,
		exchangeDeclareConfig.NoWait,
		exchangeDeclareConfig.Args)

}

func (base *QueueSetup) bindQueue() error {
	queueBindConfig := base.queueConfig.QueueBindConfig

	return base.channel.QueueBind(
		base.queueName,
		queueBindConfig.RoutingKey,
		base.exchangeName,
		queueBindConfig.NoWait,
		queueBindConfig.Args)
}
