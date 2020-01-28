package atomic

import "sync/atomic"

// Flag is an atomic flag
type Flag struct {
	value uint32
}

// Set sets flag
func (flag *Flag) Set() {
	atomic.StoreUint32(&flag.value, 1)
}

// Unset sets flag
func (flag *Flag) Unset() {
	atomic.StoreUint32(&flag.value, 0)
}

// IsSet checks whether flag is set
func (flag *Flag) IsSet() bool {
	value := atomic.LoadUint32(&flag.value)
	return value == 1
}
