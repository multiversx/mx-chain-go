package disabled

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
)

type disabledStoragePruningManager struct {
}

// NewDisabledStoragePruningManager creates a new instance of disabledStoragePruningManager
func NewDisabledStoragePruningManager() *disabledStoragePruningManager {
	return &disabledStoragePruningManager{}
}

// MarkForEviction does nothing for this implementation
func (i *disabledStoragePruningManager) MarkForEviction(_ []byte, _ []byte, _ common.ModifiedHashes, _ common.ModifiedHashes) error {
	return nil
}

// PruneTrie does nothing for this implementation
func (i *disabledStoragePruningManager) PruneTrie(_ []byte, _ state.TriePruningIdentifier, _ common.StorageManager, _ state.PruningHandler) {
}

// CancelPrune does nothing for this implementation
func (i *disabledStoragePruningManager) CancelPrune(_ []byte, _ state.TriePruningIdentifier, _ common.StorageManager) {
}

// Close does nothing for this implementation
func (i *disabledStoragePruningManager) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (i *disabledStoragePruningManager) IsInterfaceNil() bool {
	return i == nil
}
