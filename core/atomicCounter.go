package core

import "sync/atomic"

// AtomicCounter is
type AtomicCounter int64

// Set sets counter
func (counter *AtomicCounter) Set(value int64) {
	atomic.StoreInt64((*int64)(counter), value)
}

// Increment increments counter
func (counter *AtomicCounter) Increment() int64 {
	return atomic.AddInt64((*int64)(counter), 1)
}

// Decrement decrements counter
func (counter *AtomicCounter) Decrement() int64 {
	return atomic.AddInt64((*int64)(counter), -1)
}

// Get gets counter
func (counter *AtomicCounter) Get() int64 {
	return atomic.LoadInt64((*int64)(counter))
}
