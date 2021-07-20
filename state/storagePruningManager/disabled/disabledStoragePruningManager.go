package disabled

import (
	"github.com/ElrondNetwork/elrond-go/state/temporary"
)

type disabledStoragePruningManager struct {
}

// NewDisabledStoragePruningManager creates a new instance of disabledStoragePruningManager
func NewDisabledStoragePruningManager() *disabledStoragePruningManager {
	return &disabledStoragePruningManager{}
}

// MarkForEviction does nothing for this implementation
func (i *disabledStoragePruningManager) MarkForEviction(_ []byte, _ []byte, _ temporary.ModifiedHashes, _ temporary.ModifiedHashes) error {
	return nil
}

// PruneTrie does nothing for this implementation
func (i *disabledStoragePruningManager) PruneTrie(_ []byte, _ temporary.TriePruningIdentifier, _ temporary.StorageManager) {
}

// CancelPrune does nothing for this implementation
func (i *disabledStoragePruningManager) CancelPrune(_ []byte, _ temporary.TriePruningIdentifier, _ temporary.StorageManager) {
}

// Close does nothing for this implementation
func (i *disabledStoragePruningManager) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (i *disabledStoragePruningManager) IsInterfaceNil() bool {
	return i == nil
}
