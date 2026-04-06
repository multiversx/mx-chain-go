package triesHolder

import "github.com/multiversx/mx-chain-go/common"

type disabledDataTriesHolder struct{}

// NewDisabledDataTriesHolder creates a disabled no-op data tries holder.
func NewDisabledDataTriesHolder() *disabledDataTriesHolder {
	return &disabledDataTriesHolder{}
}

// Put does nothing for the disabled implementation.
func (d *disabledDataTriesHolder) Put(_ []byte, _ common.Trie) {
}

// Get always returns nil for the disabled implementation.
func (d *disabledDataTriesHolder) Get(_ []byte) common.Trie {
	return nil
}

// GetAll always returns an empty list for the disabled implementation.
func (d *disabledDataTriesHolder) GetAll() []common.Trie {
	return make([]common.Trie, 0)
}

// MarkAsDirty does nothing for the disabled implementation.
func (d *disabledDataTriesHolder) MarkAsDirty(_ []byte) {
}

// Reset does nothing for the disabled implementation.
func (d *disabledDataTriesHolder) Reset() {
}

// IsInterfaceNil returns true if the underlying object is nil.
func (d *disabledDataTriesHolder) IsInterfaceNil() bool {
	return d == nil
}
