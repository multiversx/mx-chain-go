package atomic

import "sync/atomic"

// Int64 is a wrapper for atomic operations on int64
type Int64 struct {
	value int64
}

// Set sets the value
func (variable *Int64) Set(value int64) {
	atomic.StoreInt64(&variable.value, value)
}

// Get gets the value
func (variable *Int64) Get() int64 {
	return atomic.LoadInt64(&variable.value)
}
