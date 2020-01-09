package core

import "sync/atomic"

// AtomicFlag is an atomic flag
type AtomicFlag int64

// Set sets flag
func (flag *AtomicFlag) Set() {
	atomic.StoreInt64((*int64)(flag), 1)
}

// Unset sets flag
func (flag *AtomicFlag) Unset() {
	atomic.StoreInt64((*int64)(flag), 0)
}

// IsSet checks whether flag is set
func (flag *AtomicFlag) IsSet() bool {
	value := atomic.LoadInt64((*int64)(flag))
	return value == 1
}
