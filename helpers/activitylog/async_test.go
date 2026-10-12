package activitylog

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// fakeSink records batches; block makes Publish wait, fail makes it error.
type fakeSink struct {
	mu      sync.Mutex
	batches [][]Message
	calls   int
	block   chan struct{}
	failN   int // fail the first failN calls entirely
	closed  bool
}

func (f *fakeSink) Publish(batch []Message) ([]Message, error) {
	if f.block != nil {
		<-f.block
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	if f.calls <= f.failN {
		return batch, errors.New("broker down")
	}
	f.batches = append(f.batches, append([]Message(nil), batch...))
	return nil, nil
}

func (f *fakeSink) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	return nil
}

func (f *fakeSink) total() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, b := range f.batches {
		n += len(b)
	}
	return n
}

type fallbackCollector struct {
	mu   sync.Mutex
	msgs []Message
}

func (c *fallbackCollector) add(m Message) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.msgs = append(c.msgs, m)
}

func (c *fallbackCollector) len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.msgs)
}

func newAsyncForTest(sink Sink, cfg AsyncConfig) (*Async, *fallbackCollector) {
	a := NewAsync(sink, cfg)
	fb := &fallbackCollector{}
	a.fallback = fb.add
	return a, fb
}

func shutdown(t *testing.T, a *Async) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := a.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
}

func TestAsyncBatchesAndDrainsOnShutdown(t *testing.T) {
	sink := &fakeSink{}
	a, fb := newAsyncForTest(sink, AsyncConfig{BatchSize: 10, FlushInterval: time.Hour})

	for i := 0; i < 25; i++ {
		if !a.Enqueue(Message{Key: "k", Body: []byte("x")}) {
			t.Fatal("enqueue refused")
		}
	}
	shutdown(t, a)

	if sink.total() != 25 {
		t.Fatalf("delivered %d, want 25", sink.total())
	}
	if len(sink.batches) < 3 {
		t.Errorf("batches = %d, want >= 3 (batch size 10)", len(sink.batches))
	}
	if fb.len() != 0 {
		t.Errorf("fallback got %d, want 0", fb.len())
	}
	if !sink.closed {
		t.Error("sink not closed on shutdown")
	}
}

func TestAsyncFlushIntervalSendsPartialBatch(t *testing.T) {
	sink := &fakeSink{}
	a, _ := newAsyncForTest(sink, AsyncConfig{BatchSize: 100, FlushInterval: 20 * time.Millisecond})
	defer shutdown(t, a)

	a.Enqueue(Message{Body: []byte("x")})

	deadline := time.Now().Add(time.Second)
	for sink.total() == 0 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if sink.total() != 1 {
		t.Fatalf("delivered %d, want 1 without waiting for a full batch", sink.total())
	}
}

// The whole point: a stuck broker must not block the caller. The worker is
// stuck inside Publish; Enqueue keeps returning immediately, and overflow goes
// to the fallback.
func TestAsyncSlowSinkNeverBlocksEnqueue(t *testing.T) {
	sink := &fakeSink{block: make(chan struct{})}
	a, _ := newAsyncForTest(sink, AsyncConfig{BufferSize: 5, BatchSize: 1, FlushInterval: time.Hour})

	start := time.Now()
	accepted := 0
	for i := 0; i < 1000; i++ {
		if a.Enqueue(Message{Body: []byte("x")}) {
			accepted++
		}
	}
	elapsed := time.Since(start)

	if elapsed > 500*time.Millisecond {
		t.Fatalf("1000 enqueues took %v with a stuck sink; must not block", elapsed)
	}
	if accepted >= 1000 || accepted == 0 {
		t.Fatalf("accepted %d of 1000, want bounded by the buffer", accepted)
	}
	if _, _, dropped, _ := a.Stats(); dropped != int64(1000-accepted) {
		t.Errorf("dropped = %d, want %d", dropped, 1000-accepted)
	}

	close(sink.block)
	shutdown(t, a)
}

func TestAsyncRetriesThenSucceeds(t *testing.T) {
	sink := &fakeSink{failN: 2}
	a, fb := newAsyncForTest(sink, AsyncConfig{BatchSize: 1, MaxRetries: 3, RetryBackoff: time.Millisecond})

	a.Enqueue(Message{Body: []byte("x")})
	shutdown(t, a)

	if sink.total() != 1 {
		t.Fatalf("delivered %d, want 1 after retries", sink.total())
	}
	if fb.len() != 0 {
		t.Errorf("fallback got %d, want 0", fb.len())
	}
}

func TestAsyncGivesUpToFallback(t *testing.T) {
	sink := &fakeSink{failN: 1000}
	a, fb := newAsyncForTest(sink, AsyncConfig{BatchSize: 10, MaxRetries: 2, RetryBackoff: time.Millisecond})

	a.Enqueue(Message{Body: []byte("a")})
	a.Enqueue(Message{Body: []byte("b")})
	shutdown(t, a)

	if fb.len() != 2 {
		t.Fatalf("fallback got %d, want 2 after giving up", fb.len())
	}
	if _, _, _, failed := a.Stats(); failed != 2 {
		t.Errorf("failed = %d, want 2", failed)
	}
}

func TestAsyncEnqueueAfterShutdownIsRefused(t *testing.T) {
	a, _ := newAsyncForTest(&fakeSink{}, AsyncConfig{})
	shutdown(t, a)

	if a.Enqueue(Message{Body: []byte("x")}) {
		t.Fatal("enqueue accepted after shutdown")
	}
}

type panicSink struct{ fakeSink }

func (p *panicSink) Publish([]Message) ([]Message, error) { panic("boom") }

func TestAsyncSinkPanicFallsBack(t *testing.T) {
	a, fb := newAsyncForTest(&panicSink{}, AsyncConfig{BatchSize: 1, MaxRetries: 1, RetryBackoff: time.Millisecond})

	a.Enqueue(Message{Body: []byte("x")})
	shutdown(t, a)

	if fb.len() != 1 {
		t.Fatalf("fallback got %d, want 1", fb.len())
	}
}
