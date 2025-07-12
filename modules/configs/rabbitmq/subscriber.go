package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type ConsumerHandlerData struct {
	EventType           string      `json:"event_type"` // event type name, ex : cart-item-created, user-created
	Date                string      `json:"date"`       // date publish event, ex : 2019-06-24T03:58:27.216980551Z
	Data                interface{} `json:"data"`
	CurrentExchangeName string      `json:"current_exchange_name"`
	TotalReHit          int64       `json:"total_re_hit"`
}

// ConsumerHandler function for subscribers message handler
type ConsumerHandler func(ConsumerHandlerData)

type ConsumerConfig struct {
	Consumer  string     // Consumer tag (used to identify the consumer); empty for auto-generated
	AutoAck   bool       // If true, messages are considered acknowledged once delivered
	Exclusive bool       // If true, restricts this consumer to this connection only
	NoLocal   bool       // If true, the server will not deliver messages published on this connection (not supported in RabbitMQ)
	NoWait    bool       // If true, the server does not respond to the method (fire-and-forget)
	Args      amqp.Table // Optional arguments (e.g., consumer priority)
}

func (base *QueueSetup) SetupQueue(queueDeclare *QueueDeclareConfig, consumerConfig *ConsumerConfig) *QueueSetup {
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
			Consumer:  "", // consumer tag
			AutoAck:   false,
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

	if !consumerConfig.AutoAck && base.channel != nil {
		loggingMessage("setting up qos", nil)
		err := base.channel.Qos(1, 0, false)
		if err != nil {
			loggingMessage("error setting up QoS", err.Error())
		}
	}

	return base
}

func (base *QueueSetup) SetupQueueBind(queueBindConfig *QueueBindConfig) *QueueSetup {
	if queueBindConfig == nil {
		queueBindConfig = &QueueBindConfig{
			RoutingKey: "",
			NoWait:     false,
			Args:       nil,
		}
	}

	base.queueConfig.QueueBindConfig = queueBindConfig
	return base
}

func (base *QueueSetup) AddConsumer(isReconnect bool) *QueueSetup {
	err := base.openConnection()
	if err != nil {
		loggingMessage("error open new connection", err.Error())
		panic(err.Error())
	}

	err = base.declareQueue()
	if err != nil {
		loggingMessage("error declare queue after open connection", err.Error())
		panic(err.Error())
	}

	if !isReconnect {
		go base.reconnect()
	}

	return base
}

func (base *QueueSetup) Consume(consumer ConsumerHandler) {
	loggingMessage("Registering Consumer...", nil)
	deliveries, err := base.registerQueueConsumer()
	if err != nil {
		loggingMessage("Error register queue queueConsumer", err.Error())
		panic(err.Error())
	}

	base.executeMessageConsumer(consumer, deliveries, false)
}
