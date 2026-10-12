package activitylog

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"time"

	"github.com/IBM/sarama"
	"github.com/fari-99/go-helper/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// output is the active asynchronous pipeline; nil means "write to stdout".
var output atomic.Pointer[Async]

// Configure routes every activity log through an asynchronous pipeline in front
// of sink instead of writing to stdout. Call it once at startup; call Shutdown
// before the process exits so queued logs are flushed.
//
// Replacing an existing pipeline shuts the old one down in the background.
func Configure(sink Sink, cfg AsyncConfig) *Async {
	next := NewAsync(sink, cfg)
	if prev := output.Swap(next); prev != nil {
		go func() { _ = prev.Shutdown(context.Background()) }()
	}
	return next
}

// Shutdown flushes and stops the configured pipeline (no-op when logging to
// stdout) and goes back to stdout.
func Shutdown(ctx context.Context) error {
	if a := output.Swap(nil); a != nil {
		return a.Shutdown(ctx)
	}
	return nil
}

// deliver hands a marshaled log to the pipeline, or to stdout when there is no
// pipeline or it cannot take the message right now.
func deliver(m Message) {
	if a := output.Load(); a != nil && a.Enqueue(m) {
		return
	}
	writeStdout(m)
}

// writeStdout is the default (and fallback) destination.
func writeStdout(m Message) {
	log.Print(string(m.Body))
}

// KafkaSink publishes batches with one SendMessages round trip each.
type KafkaSink struct {
	producer sarama.SyncProducer
	topic    string
}

// NewKafkaSink wraps an existing producer. Ownership moves to the sink: Close
// closes the producer.
func NewKafkaSink(producer sarama.SyncProducer, topic string) *KafkaSink {
	return &KafkaSink{producer: producer, topic: topic}
}

func (k *KafkaSink) Publish(batch []Message) ([]Message, error) {
	msgs := make([]*sarama.ProducerMessage, len(batch))
	for i, m := range batch {
		msgs[i] = &sarama.ProducerMessage{
			Topic:    k.topic,
			Key:      sarama.StringEncoder(m.Key), // same entity -> same partition -> ordered
			Value:    sarama.ByteEncoder(m.Body),
			Metadata: i,
		}
	}

	err := k.producer.SendMessages(msgs)
	if err == nil {
		return nil, nil
	}

	// Per-message errors say exactly which ones failed; anything else (e.g. the
	// whole call failed) means we cannot tell, so retry the batch.
	producerErrs, ok := err.(sarama.ProducerErrors)
	if !ok {
		return batch, err
	}

	failed := make([]Message, 0, len(producerErrs))
	for _, pe := range producerErrs {
		if idx, ok := pe.Msg.Metadata.(int); ok && idx < len(batch) {
			failed = append(failed, batch[idx])
		}
	}
	if len(failed) == 0 {
		return batch, err
	}
	return failed, err
}

func (k *KafkaSink) Close() error { return k.producer.Close() }

// RabbitMQSink publishes each message to a queue. The helper's publisher is not
// safe to share and its reconnect logic gives up after a few attempts, so the
// sink owns it: on any publish error the publisher is closed and rebuilt on
// the next attempt.
type RabbitMQSink struct {
	queueName string
	publisher *rabbitmq.QueueSetup
}

// NewRabbitMQSink takes an already connected publisher (from
// NewRabbitMQPublisher; may be nil to connect on first use) for queueName.
func NewRabbitMQSink(queueName string, publisher *rabbitmq.QueueSetup) *RabbitMQSink {
	return &RabbitMQSink{queueName: queueName, publisher: publisher}
}

// NewRabbitMQPublisher builds a durable, persistent publisher for queueName,
// turning the helper's panics into errors.
func NewRabbitMQPublisher(queueName string) (publisher *rabbitmq.QueueSetup, err error) {
	defer func() {
		if r := recover(); r != nil {
			publisher, err = nil, fmt.Errorf("rabbitmq publisher: %v", r)
		}
	}()

	// The helper dials in an endless loop when the broker is down, which would
	// hang startup and the sink's worker. Fail fast with a bounded probe first.
	if err := probeRabbitMQ(); err != nil {
		return nil, fmt.Errorf("rabbitmq publisher: %w", err)
	}

	queue := rabbitmq.NewBaseQueue("", queueName)
	if queue == nil {
		return nil, fmt.Errorf("rabbitmq publisher: could not open connection")
	}

	queue.AddPublisher(
		&rabbitmq.QueueDeclareConfig{Durable: true},
		&rabbitmq.PublisherConfig{
			Msg: amqp.Publishing{
				ContentType:  "application/json",
				DeliveryMode: amqp.Persistent,
			},
		},
	)
	return queue, nil
}

// probeRabbitMQ checks the broker is reachable (same RABBIT_* settings the
// helper uses) within a few seconds.
func probeRabbitMQ() error {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		os.Getenv("RABBIT_USER"), os.Getenv("RABBIT_PASSWORD"),
		os.Getenv("RABBIT_HOST"), os.Getenv("RABBIT_PORT"))

	conn, err := amqp.DialConfig(url, amqp.Config{Dial: amqp.DefaultDial(5 * time.Second)})
	if err != nil {
		return err
	}
	return conn.Close()
}

func (r *RabbitMQSink) Publish(batch []Message) ([]Message, error) {
	if r.publisher == nil {
		publisher, err := NewRabbitMQPublisher(r.queueName)
		if err != nil {
			return batch, err
		}
		r.publisher = publisher
	}

	for i, m := range batch {
		if err := r.publisher.Publish(string(m.Body)); err != nil {
			r.reset()
			return batch[i:], err // [:i] were published
		}
	}
	return nil, nil
}

func (r *RabbitMQSink) Close() error {
	r.reset()
	return nil
}

func (r *RabbitMQSink) reset() {
	if r.publisher != nil {
		r.publisher.Close()
		r.publisher = nil
	}
}
