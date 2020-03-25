package atomic

import "sync/atomic"

// String is a wrapper for atomic operations on String
type String struct {
	value atomic.Value
}

// Set sets the value
func (variable *String) Set(value string) {
	variable.value.Store(value)
}

// Get gets the value
func (variable *String) Get() string {
	value := variable.value.Load()
	if value == nil {
		return ""
	}

	return value.(string)
}
