package configs

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/fari-99/go-helper/rabbitmq"
	"github.com/spf13/cast"

	"go-api/helpers/activitylog"
	"go-api/modules/configs/kafka"
)

// Activity log destinations, selected with ACTIVITY_LOG_DRIVER.
const (
	ActivityLogDriverStdout   = "stdout"
	ActivityLogDriverRabbitMQ = "rabbitmq"
	ActivityLogDriverKafka    = "kafka"
)

// InitActivityLog routes activity logs to the driver named by
// ACTIVITY_LOG_DRIVER (stdout when unset). Broker drivers sit behind an
// in-memory queue (see activitylog.Async) so a slow or unreachable broker never
// slows a database operation. If the broker cannot be reached at startup the
// service keeps running and logs to stdout.
//
// Call once at startup, and activitylog.Shutdown before exit to flush.
func InitActivityLog() {
	driver := strings.ToLower(strings.TrimSpace(os.Getenv("ACTIVITY_LOG_DRIVER")))

	var (
		sink activitylog.Sink
		err  error
	)

	switch driver {
	case "", ActivityLogDriverStdout:
		return

	case ActivityLogDriverRabbitMQ:
		queue := envOrDefault("ACTIVITY_LOG_RABBITMQ_QUEUE", "activity_logs")
		// Connect eagerly so a bad config shows up at boot, not on the first write.
		var publisher *rabbitmq.QueueSetup
		publisher, err = activitylog.NewRabbitMQPublisher(queue)
		if err == nil {
			sink = activitylog.NewRabbitMQSink(queue, publisher)
		}

	case ActivityLogDriverKafka:
		var producer sarama.SyncProducer
		producer, err = newKafkaProducer()
		if err == nil {
			sink = activitylog.NewKafkaSink(producer, envOrDefault("ACTIVITY_LOG_KAFKA_TOPIC", "activity_logs"))
		}

	default:
		log.Printf("activitylog: unknown ACTIVITY_LOG_DRIVER %q, using stdout", driver)
		return
	}

	if err != nil {
		log.Printf("activitylog: %s driver unavailable, using stdout: %v", driver, err)
		return
	}

	activitylog.Configure(sink, activitylog.AsyncConfig{
		BufferSize:    cast.ToInt(os.Getenv("ACTIVITY_LOG_BUFFER_SIZE")),
		BatchSize:     cast.ToInt(os.Getenv("ACTIVITY_LOG_BATCH_SIZE")),
		FlushInterval: time.Duration(cast.ToInt(os.Getenv("ACTIVITY_LOG_FLUSH_INTERVAL_MS"))) * time.Millisecond,
		MaxRetries:    cast.ToInt(os.Getenv("ACTIVITY_LOG_MAX_RETRIES")),
	})
	log.Printf("activitylog: publishing to %s", driver)
}

func envOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

// newKafkaProducer wraps the shared producer getter, which panics when the
// broker is unreachable.
func newKafkaProducer() (producer sarama.SyncProducer, err error) {
	defer func() {
		if r := recover(); r != nil {
			producer, err = nil, fmt.Errorf("kafka producer: %v", r)
		}
	}()
	return kafka.GetKafkaProducerConnection(), nil
}
