package disabled

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
)

type disabledSnapshotsManger struct {
}

// NewDisabledSnapshotsManager creates a new disabled snapshots manager
func NewDisabledSnapshotsManager() state.SnapshotsManager {
	return &disabledSnapshotsManger{}
}

// SnapshotState does nothing for this implementation
func (d *disabledSnapshotsManger) SnapshotState(_ []byte, _ uint32, _ common.StorageManager) {
}

// StartSnapshotAfterRestartIfNeeded returns nil for this implementation
func (d *disabledSnapshotsManger) StartSnapshotAfterRestartIfNeeded(_ common.StorageManager) error {
	return nil
}

// IsSnapshotInProgress returns false for this implementation
func (d *disabledSnapshotsManger) IsSnapshotInProgress() bool {
	return false
}

// SetSyncer returns nil for this implementation
func (d *disabledSnapshotsManger) SetSyncer(_ state.AccountsDBSyncer) error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (d *disabledSnapshotsManger) IsInterfaceNil() bool {
	return d == nil
}
