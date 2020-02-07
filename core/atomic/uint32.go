package atomic

import "sync/atomic"

// Uint32 is a wrapper for atomic operations on uint32
type Uint32 struct {
	value uint32
}

// Set sets the value
func (variable *Uint32) Set(value uint32) {
	atomic.StoreUint32(&variable.value, value)
}

// Get gets the value
func (variable *Uint32) Get() uint32 {
	return atomic.LoadUint32(&variable.value)
}
