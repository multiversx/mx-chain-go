package core

import "sync/atomic"

// AtomicCounter is an atomic counter
type AtomicCounter int64

// AtomicUint32 is a wrapper for atomic operations on uint32
type AtomicUint32 uint32

// Set sets counter
func (counter *AtomicCounter) Set(value int64) {
	atomic.StoreInt64((*int64)(counter), value)
}

// Increment increments counter
func (counter *AtomicCounter) Increment() int64 {
	return atomic.AddInt64((*int64)(counter), 1)
}

// Add adds value to counter
func (counter *AtomicCounter) Add(value int64) int64 {
	return atomic.AddInt64((*int64)(counter), value)
}

// Decrement decrements counter
func (counter *AtomicCounter) Decrement() int64 {
	return atomic.AddInt64((*int64)(counter), -1)
}

// Subtract subtracts value from counter
func (counter *AtomicCounter) Subtract(value int64) int64 {
	return atomic.AddInt64((*int64)(counter), -value)
}

// Get gets counter
func (counter *AtomicCounter) Get() int64 {
	return atomic.LoadInt64((*int64)(counter))
}

// GetUint64 gets counter as uint64
func (counter *AtomicCounter) GetUint64() uint64 {
	value := counter.Get()
	if value < 0 {
		return 0
	}
	return uint64(value)
}

// Set sets the value
func (variable *AtomicUint32) Set(value uint32) {
	atomic.StoreUint32((*uint32)(variable), value)
}

// Get gets the value
func (variable *AtomicUint32) Get() uint32 {
	return atomic.LoadUint32((*uint32)(variable))
}
