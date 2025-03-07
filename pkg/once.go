package pkg

import (
	"sync"
	"sync/atomic"
)

// Once All of this function are from sync.Once
type Once struct {
	done atomic.Uint32
	m    sync.Mutex
}

func (o *Once) Do(f func()) {
	if o.done.Load() == 0 {
		o.doSlow(f)
	}
}

func (o *Once) doSlow(f func()) {
	o.m.Lock()
	defer o.m.Unlock()
	if o.done.Load() == 0 {
		defer o.done.Store(1)
		f()
	}
}

// Reset set async Once.done into 0
func (o *Once) Reset() {
	o.m.Lock()
	defer o.m.Unlock()

	o.done.Store(0)
}
