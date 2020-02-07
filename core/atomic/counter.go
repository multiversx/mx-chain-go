package atomic

import "sync/atomic"

// Counter is an atomic counter
type Counter struct {
	value int64
}

// Set sets counter
func (counter *Counter) Set(value int64) {
	atomic.StoreInt64(&counter.value, value)
}

// Increment increments counter
func (counter *Counter) Increment() int64 {
	return atomic.AddInt64(&counter.value, 1)
}

// Add adds value to counter
func (counter *Counter) Add(value int64) int64 {
	return atomic.AddInt64(&counter.value, value)
}

// Decrement decrements counter
func (counter *Counter) Decrement() int64 {
	return atomic.AddInt64(&counter.value, -1)
}

// Subtract subtracts value from counter
func (counter *Counter) Subtract(value int64) int64 {
	return atomic.AddInt64(&counter.value, -value)
}

// Get gets counter
func (counter *Counter) Get() int64 {
	return atomic.LoadInt64(&counter.value)
}

// Reset resets counter and returns the previous value
func (counter *Counter) Reset() int64 {
	return atomic.SwapInt64(&counter.value, 0)
}

// GetUint64 gets counter as uint64
func (counter *Counter) GetUint64() uint64 {
	value := counter.Get()
	if value < 0 {
		return 0
	}
	return uint64(value)
}
