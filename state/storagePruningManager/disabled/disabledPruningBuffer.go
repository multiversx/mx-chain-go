package disabled

import "math"

type disabledPruningBuffer struct {
}

// NewDisabledPruningBuffer creates a new instance of disabledPruningBuffer
func NewDisabledPruningBuffer() *disabledPruningBuffer {
	return &disabledPruningBuffer{}
}

// Add does nothing
func (dpb *disabledPruningBuffer) Add(_ []byte) {
}

// RemoveAll returns an empty slice
func (dpb *disabledPruningBuffer) RemoveAll() [][]byte {
	return make([][]byte, 0)
}

// Len returns 0
func (dpb *disabledPruningBuffer) Len() int {
	return 0
}

// MaximumSize returns max int 32
func (dpb *disabledPruningBuffer) MaximumSize() int {
	return math.MaxInt32
}

// IsInterfaceNil returns true if there is no value under the interface
func (dpb *disabledPruningBuffer) IsInterfaceNil() bool {
	return dpb == nil
}
