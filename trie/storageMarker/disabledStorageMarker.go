package storageMarker

import "github.com/multiversx/mx-chain-go/common"

type disabledStorageMarker struct {
}

// NewDisabledStorageMarker creates a new instance of disabledStorageMarker
func NewDisabledStorageMarker() *disabledStorageMarker {
	return &disabledStorageMarker{}
}

// MarkStorerAsSyncedAndActive does nothing for this implementation
func (dsm *disabledStorageMarker) MarkStorerAsSyncedAndActive(_ common.StorageManager) {
}
