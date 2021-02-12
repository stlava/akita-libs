package batcher

import (
	"sync"
	"time"
)

type InMemory struct {
	processor  BatchProcessor
	buf        []interface{} // protected by mu
	maxBufSize int

	signalClose                chan struct{} // closed when the batcher is closed
	signalPeriodicFlushStopped chan struct{} // closed when we stopped periodic flushing
	mu                         sync.Mutex
}

func NewInMemory(p BatchProcessor, maxBufSize int, flushDuration time.Duration) *InMemory {
	m := &InMemory{
		processor:                  p,
		buf:                        make([]interface{}, 0, maxBufSize),
		maxBufSize:                 maxBufSize,
		signalClose:                make(chan struct{}),
		signalPeriodicFlushStopped: make(chan struct{}),
	}

	go func() {
		defer close(m.signalPeriodicFlushStopped)
		ticker := time.NewTicker(flushDuration)
		defer ticker.Stop()
		for {
			select {
			case <-m.signalClose:
				return
			case <-ticker.C:
				m.mu.Lock()
				m.flush()
				m.mu.Unlock()
			}
		}
	}()

	return m
}

func (m *InMemory) Add(items ...interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// TODO: check if appending all items will put us over maxBufSize
	m.buf = append(m.buf, items...)
	if len(m.buf) >= m.maxBufSize {
		m.flush()
	}
	return nil
}

func (m *InMemory) Close() {
	// Stop the periodic flusher.
	close(m.signalClose)
	<-m.signalPeriodicFlushStopped

	m.mu.Lock()
	defer m.mu.Unlock()
	m.flush()
}

// Must hold mu when calling flush
func (m *InMemory) flush() {
	if len(m.buf) == 0 {
		return
	}

	m.processor(m.buf)
	// Clear buf without reallocating memory.
	m.buf = m.buf[:0]
}
