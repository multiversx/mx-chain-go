package disabled

import "github.com/multiversx/mx-chain-go/common"

type disabledDataTriesHolder struct {
}

// NewDisabledDataTriesHolder creates a new disabled data tries holder.
func NewDisabledDataTriesHolder() common.TriesHolder {
	return &disabledDataTriesHolder{}
}

// Put does nothing for this implementation.
func (ddth *disabledDataTriesHolder) Put(_ []byte, _ common.Trie) {
}

// Get returns nil for this implementation.
func (ddth *disabledDataTriesHolder) Get(_ []byte) common.Trie {
	return nil
}

// GetAll returns an empty slice for this implementation.
func (ddth *disabledDataTriesHolder) GetAll() []common.Trie {
	return make([]common.Trie, 0)
}

// MarkAsDirty does nothing for this implementation.
func (ddth *disabledDataTriesHolder) MarkAsDirty(_ []byte) {
}

// Reset does nothing for this implementation.
func (ddth *disabledDataTriesHolder) Reset() {
}

// IsInterfaceNil returns true if there is no value under the interface.
func (ddth *disabledDataTriesHolder) IsInterfaceNil() bool {
	return ddth == nil
}
