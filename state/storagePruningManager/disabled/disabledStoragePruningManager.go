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
func (d *disabledStoragePruningManager) MarkForEviction(_ []byte, _ []byte, _ common.ModifiedHashes, _ common.ModifiedHashes) error {
	return nil
}

// PruneTrie does nothing for this implementation
func (d *disabledStoragePruningManager) PruneTrie(_ []byte, _ state.TriePruningIdentifier, _ common.StorageManager, _ state.PruningHandler) {
}

// CancelPrune does nothing for this implementation
func (d *disabledStoragePruningManager) CancelPrune(_ []byte, _ state.TriePruningIdentifier, _ common.StorageManager) {
}

// Reset does nothing for this implementation
func (d *disabledStoragePruningManager) Reset() {}

// Close does nothing for this implementation
func (d *disabledStoragePruningManager) Close() error {
	return nil
}

// EvictionWaitingListCacheLen returns 0 for the disabled storage pruning manager
func (d *disabledStoragePruningManager) EvictionWaitingListCacheLen() int {
	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledStoragePruningManager) IsInterfaceNil() bool {
	return d == nil
}
