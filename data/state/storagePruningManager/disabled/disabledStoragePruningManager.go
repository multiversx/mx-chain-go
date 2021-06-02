package disabled

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type disabledStoragePruningManager struct {
}

// NewDisabledStoragePruningManager creates a new instance of disabledStoragePruningManager
func NewDisabledStoragePruningManager() *disabledStoragePruningManager {
	return &disabledStoragePruningManager{}
}

// MarkForEviction does nothing for this implementation
func (i *disabledStoragePruningManager) MarkForEviction(_ []byte, _ []byte, _ data.ModifiedHashes, _ data.ModifiedHashes) error {
	return nil
}

// PruneTrie does nothing for this implementation
func (i *disabledStoragePruningManager) PruneTrie(_ []byte, _ data.TriePruningIdentifier, _ data.StorageManager) {
}

// CancelPrune does nothing for this implementation
func (i *disabledStoragePruningManager) CancelPrune(_ []byte, _ data.TriePruningIdentifier, _ data.StorageManager) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (i *disabledStoragePruningManager) IsInterfaceNil() bool {
	return i == nil
}
