package storagePruningManager

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type inactiveStoragePruningManager struct {
}

// NewInactiveStoragePruningManager creates a new instance of inactiveStoragePruningManager
func NewInactiveStoragePruningManager() *inactiveStoragePruningManager {
	return &inactiveStoragePruningManager{}
}

// MarkForEviction does nothing for this implementation
func (i *inactiveStoragePruningManager) MarkForEviction(_ []byte, _ []byte, _ data.ModifiedHashes, _ data.ModifiedHashes) error {
	return nil
}

// PruneTrie does nothing for this implementation
func (i *inactiveStoragePruningManager) PruneTrie(_ []byte, _ data.TriePruningIdentifier, _ data.StorageManager) {
}

// CancelPrune does nothing for this implementation
func (i *inactiveStoragePruningManager) CancelPrune(_ []byte, _ data.TriePruningIdentifier, _ data.StorageManager) {
}

// IsInterfaceNil returns true if there is no value under the interface
func (i *inactiveStoragePruningManager) IsInterfaceNil() bool {
	return i == nil
}
