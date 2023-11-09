package disabled

import (
	"github.com/multiversx/mx-chain-go/common"
)

type disabledCheckpointHashesHolder struct {
}

// NewDisabledCheckpointHashesHolder creates a new instance of disabledCheckpointHashesHolder
func NewDisabledCheckpointHashesHolder() *disabledCheckpointHashesHolder {
	return &disabledCheckpointHashesHolder{}
}

// Put returns false
func (d *disabledCheckpointHashesHolder) Put(_ []byte, _ common.ModifiedHashes) bool {
	return false
}

// RemoveCommitted does nothing for this implementation
func (d *disabledCheckpointHashesHolder) RemoveCommitted(_ []byte) {
}

// Remove does nothing for this implementation
func (d *disabledCheckpointHashesHolder) Remove(_ []byte) {
}

// ShouldCommit returns true
func (d *disabledCheckpointHashesHolder) ShouldCommit(_ []byte) bool {
	return true
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledCheckpointHashesHolder) IsInterfaceNil() bool {
	return d == nil
}
