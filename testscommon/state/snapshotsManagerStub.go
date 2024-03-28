package state

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
)

// SnapshotsManagerStub -
type SnapshotsManagerStub struct {
	SnapshotStateCalled                     func(rootHash []byte, epoch uint32, trieStorageManager common.StorageManager)
	StartSnapshotAfterRestartIfNeededCalled func(trieStorageManager common.StorageManager) error
	IsSnapshotInProgressCalled              func() bool
	SetSyncerCalled                         func(syncer state.AccountsDBSyncer) error
}

// SnapshotState -
func (s *SnapshotsManagerStub) SnapshotState(rootHash []byte, epoch uint32, trieStorageManager common.StorageManager) {
	if s.SnapshotStateCalled != nil {
		s.SnapshotStateCalled(rootHash, epoch, trieStorageManager)
	}
}

// StartSnapshotAfterRestartIfNeeded -
func (s *SnapshotsManagerStub) StartSnapshotAfterRestartIfNeeded(trieStorageManager common.StorageManager) error {
	if s.StartSnapshotAfterRestartIfNeededCalled != nil {
		return s.StartSnapshotAfterRestartIfNeededCalled(trieStorageManager)
	}
	return nil
}

// IsSnapshotInProgress -
func (s *SnapshotsManagerStub) IsSnapshotInProgress() bool {
	if s.IsSnapshotInProgressCalled != nil {
		return s.IsSnapshotInProgressCalled()
	}
	return false
}

// SetSyncer -
func (s *SnapshotsManagerStub) SetSyncer(syncer state.AccountsDBSyncer) error {
	if s.SetSyncerCalled != nil {
		return s.SetSyncerCalled(syncer)
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *SnapshotsManagerStub) IsInterfaceNil() bool {
	return s == nil
}
