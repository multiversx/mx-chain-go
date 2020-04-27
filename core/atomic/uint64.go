package atomic

import "sync/atomic"

// Uint64 is a wrapper for atomic operations on uint64
type Uint64 struct {
	value uint64
}

// Set sets the value
func (variable *Uint64) Set(value uint64) {
	atomic.StoreUint64(&variable.value, value)
}

// Get gets the value
func (variable *Uint64) Get() uint64 {
	return atomic.LoadUint64(&variable.value)
}
