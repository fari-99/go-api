package activitylog

import (
	"context"
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// Message is one marshaled activity log on its way to a Sink.
type Message struct {
	// Key groups messages that must stay ordered relative to each other (the
	// entity and its id); sinks that support partitioning use it.
	Key  string
	Body []byte
}

// Sink is a destination for activity logs (RabbitMQ, Kafka, ...).
//
// Publish is always called from a single worker goroutine, so implementations
// need not be safe for concurrent use. It may block on the broker — that is
// the point of the queue in front of it. It returns the messages that were NOT
// delivered (so only those are retried) and the reason.
type Sink interface {
	Publish(batch []Message) (failed []Message, err error)
	Close() error
}

// AsyncConfig tunes the in-memory queue in front of a Sink.
type AsyncConfig struct {
	// BufferSize is how many messages may wait in memory. When it is full new
	// messages are NOT queued (they go to the stdout fallback) so the caller —
	// a database operation — never blocks. Default 10000.
	BufferSize int
	// BatchSize is the most messages handed to Sink.Publish at once. Default 100.
	BatchSize int
	// FlushInterval is how long a partial batch may wait. Default 500ms.
	FlushInterval time.Duration
	// MaxRetries is how many times a failed batch is retried (with exponential
	// backoff starting at RetryBackoff) before falling back to stdout. Default 3.
	MaxRetries int
	// RetryBackoff is the first retry delay; it doubles each attempt. Default 200ms.
	RetryBackoff time.Duration
}

func (c AsyncConfig) withDefaults() AsyncConfig {
	if c.BufferSize <= 0 {
		c.BufferSize = 10000
	}
	if c.BatchSize <= 0 {
		c.BatchSize = 100
	}
	if c.FlushInterval <= 0 {
		c.FlushInterval = 500 * time.Millisecond
	}
	if c.MaxRetries < 0 {
		c.MaxRetries = 0
	} else if c.MaxRetries == 0 {
		c.MaxRetries = 3
	}
	if c.RetryBackoff <= 0 {
		c.RetryBackoff = 200 * time.Millisecond
	}
	return c
}

// Async decouples emitting a log from delivering it. Enqueue never blocks: it
// is a non-blocking send onto a bounded channel drained by one worker, which
// batches messages, publishes them and retries. Whatever cannot be delivered
// (queue full, broker down, shutting down) is written to the fallback — stdout
// by default — instead of being dropped silently or stalling the caller.
type Async struct {
	sink     Sink
	cfg      AsyncConfig
	fallback func(Message)

	queue chan Message
	stop  chan struct{}
	done  chan struct{}

	// mu makes "check closed, then send" atomic with respect to Shutdown, so a
	// message can never be queued after the worker did its final drain.
	mu     sync.RWMutex
	closed bool

	dropped   atomic.Int64 // queue full or shut down
	failed    atomic.Int64 // publish gave up after retries
	delivered atomic.Int64
}

// NewAsync starts the worker. The fallback defaults to the stdout writer.
func NewAsync(sink Sink, cfg AsyncConfig) *Async {
	cfg = cfg.withDefaults()
	a := &Async{
		sink:     sink,
		cfg:      cfg,
		fallback: writeStdout,
		queue:    make(chan Message, cfg.BufferSize),
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
	go a.run()
	return a
}

// Enqueue queues a message without ever blocking. It reports false when the
// message was not queued (the caller should use the fallback).
func (a *Async) Enqueue(m Message) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.closed {
		a.dropped.Add(1)
		return false
	}

	select {
	case a.queue <- m:
		return true
	default:
		a.dropped.Add(1)
		return false
	}
}

// Stats reports counters since start, useful for a health/metrics endpoint.
func (a *Async) Stats() (queued, delivered, dropped, failed int64) {
	return int64(len(a.queue)), a.delivered.Load(), a.dropped.Load(), a.failed.Load()
}

// Shutdown stops accepting messages, drains what is queued and closes the sink.
// It returns ctx.Err() if the drain did not finish in time; whatever was still
// queued is then written to the fallback so nothing is lost silently.
func (a *Async) Shutdown(ctx context.Context) error {
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return nil
	}
	a.closed = true
	a.mu.Unlock()

	close(a.stop)

	select {
	case <-a.done:
		return nil
	case <-ctx.Done():
		a.flushQueueToFallback()
		return ctx.Err()
	}
}

func (a *Async) run() {
	defer close(a.done)
	defer func() {
		if err := a.sink.Close(); err != nil {
			log.Printf("activitylog: closing sink: %v", err)
		}
	}()

	batch := make([]Message, 0, a.cfg.BatchSize)
	timer := time.NewTimer(a.cfg.FlushInterval)
	defer timer.Stop()

	flush := func(stopping bool) {
		if len(batch) == 0 {
			return
		}
		a.publish(batch, stopping)
		batch = batch[:0]
	}

	for {
		select {
		case m := <-a.queue:
			batch = append(batch, m)
			if len(batch) >= a.cfg.BatchSize {
				flush(false)
			}

		case <-timer.C:
			flush(false)
			timer.Reset(a.cfg.FlushInterval)

		case <-a.stop:
			// Enqueue is closed, so the queue only shrinks from here.
			for {
				select {
				case m := <-a.queue:
					batch = append(batch, m)
					if len(batch) >= a.cfg.BatchSize {
						flush(true)
					}
				default:
					flush(true)
					return
				}
			}
		}
	}
}

// publish delivers a batch, retrying only the messages that failed. Anything
// still failing after the retries goes to the fallback.
func (a *Async) publish(batch []Message, stopping bool) {
	pending := append([]Message(nil), batch...)
	backoff := a.cfg.RetryBackoff

	for attempt := 0; ; attempt++ {
		failed, err := a.safePublish(pending)
		a.delivered.Add(int64(len(pending) - len(failed)))
		pending = failed

		if len(pending) == 0 {
			return
		}
		if attempt >= a.cfg.MaxRetries || stopping {
			log.Printf("activitylog: giving up on %d message(s), writing to fallback: %v", len(pending), err)
			a.failed.Add(int64(len(pending)))
			for _, m := range pending {
				a.fallback(m)
			}
			return
		}

		time.Sleep(backoff)
		backoff *= 2
	}
}

// safePublish turns a panic inside a sink into a failed batch.
func (a *Async) safePublish(batch []Message) (failed []Message, err error) {
	defer func() {
		if r := recover(); r != nil {
			failed = batch
			err = errors.New("sink panicked")
			log.Printf("activitylog: recovered from panic in sink: %v", r)
		}
	}()
	return a.sink.Publish(batch)
}

func (a *Async) flushQueueToFallback() {
	for {
		select {
		case m := <-a.queue:
			a.fallback(m)
		default:
			return
		}
	}
}
