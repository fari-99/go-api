package file_managements

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/fari-99/go-helper/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"

	"go-api/constant"
)

// Publisher queues a job for the worker.
type Publisher interface {
	Publish(fileManagementID uint64) error
}

type QueuePayload struct {
	FileManagementID uint64 `json:"file_management_id"`
}

// QueueDeclareConfig is shared by the publisher and the worker: both must declare the queue
// the same way. Durable, so queued jobs survive a broker restart.
func QueueDeclareConfig() *rabbitmq.QueueDeclareConfig {
	return &rabbitmq.QueueDeclareConfig{Durable: true}
}

type rabbitPublisher struct {
	mu     sync.Mutex
	queue  *rabbitmq.QueueSetup
	config *rabbitmq.PublisherConfig
}

// NewRabbitPublisher opens one long-lived connection to file-management-queue. It blocks until
// RabbitMQ is reachable, like the other queue setups.
func NewRabbitPublisher() Publisher {
	config := &rabbitmq.PublisherConfig{
		Msg: amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			AppId:        os.Getenv("APP_NAME"),
		},
	}

	queue := rabbitmq.NewBaseQueue("", constant.QueueFileManagement)
	queue.AddPublisher(QueueDeclareConfig(), config)

	return &rabbitPublisher{queue: queue, config: config}
}

func (p *rabbitPublisher) Publish(fileManagementID uint64) error {
	now := time.Now().UTC()
	message, err := json.Marshal(rabbitmq.ConsumerHandlerData{
		EventType: constant.EventFileManagementProcess,
		Date:      now.Format(time.RFC3339Nano),
		Data:      QueuePayload{FileManagementID: fileManagementID},
	})
	if err != nil {
		return err
	}

	// QueueSetup.Publish writes the body into the shared publisher config, so serialize publishes
	p.mu.Lock()
	defer p.mu.Unlock()

	p.config.Msg.Timestamp = now
	return p.queue.Publish(string(message))
}
