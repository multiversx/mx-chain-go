package txcache

import "sync/atomic"

// AtomicCounter is
type AtomicCounter int64

// Increment increments
func (counter *AtomicCounter) Increment() int64 {
	return atomic.AddInt64((*int64)(counter), 1)
}

// Decrement decrements
func (counter *AtomicCounter) Decrement() int64 {
	return atomic.AddInt64((*int64)(counter), -1)
}

// Get gets
func (counter *AtomicCounter) Get() int64 {
	return atomic.LoadInt64((*int64)(counter))
}

// GetUnstable gets
func (counter *AtomicCounter) GetUnstable() int64 {
	return (int64)(*counter)
}
