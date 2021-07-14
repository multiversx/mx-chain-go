package testscommon

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
)

// StorageManagerStub -
type StorageManagerStub struct {
	DatabaseCalled                    func() data.DBWriteCacher
	TakeSnapshotCalled                func([]byte, bool, chan core.KeyValueHolder)
	SetCheckpointCalled               func([]byte, chan core.KeyValueHolder)
	GetDbThatContainsHashCalled       func([]byte) data.DBWriteCacher
	GetSnapshotThatContainsHashCalled func(rootHash []byte) data.SnapshotDbHandler
	IsPruningEnabledCalled            func() bool
	IsPruningBlockedCalled            func() bool
	EnterPruningBufferingModeCalled   func()
	ExitPruningBufferingModeCalled    func()
	AddDirtyCheckpointHashesCalled    func([]byte, data.ModifiedHashes) bool
	RemoveCalled                      func([]byte) error
	IsInterfaceNilCalled              func() bool
}

// Database -
func (sms *StorageManagerStub) Database() data.DBWriteCacher {
	if sms.DatabaseCalled != nil {
		return sms.DatabaseCalled()
	}
	return nil
}

// TakeSnapshot -
func (sms *StorageManagerStub) TakeSnapshot(rootHash []byte, newDB bool, leavesChan chan core.KeyValueHolder) {
	if sms.TakeSnapshotCalled != nil {
		sms.TakeSnapshotCalled(rootHash, newDB, leavesChan)
	}
}

// SetCheckpoint -
func (sms *StorageManagerStub) SetCheckpoint(rootHash []byte, leavesChan chan core.KeyValueHolder) {
	if sms.SetCheckpointCalled != nil {
		sms.SetCheckpointCalled(rootHash, leavesChan)
	}
}

// GetSnapshotThatContainsHash -
func (sms *StorageManagerStub) GetSnapshotThatContainsHash(d []byte) data.SnapshotDbHandler {
	if sms.GetSnapshotThatContainsHashCalled != nil {
		return sms.GetSnapshotThatContainsHashCalled(d)
	}

	return nil
}

// IsPruningEnabled -
func (sms *StorageManagerStub) IsPruningEnabled() bool {
	if sms.IsPruningEnabledCalled != nil {
		return sms.IsPruningEnabledCalled()
	}
	return false
}

// IsPruningBlocked -
func (sms *StorageManagerStub) IsPruningBlocked() bool {
	if sms.IsPruningBlockedCalled != nil {
		return sms.IsPruningBlockedCalled()
	}
	return false
}

// EnterPruningBufferingMode -
func (sms *StorageManagerStub) EnterPruningBufferingMode() {
	if sms.EnterPruningBufferingModeCalled != nil {
		sms.EnterPruningBufferingModeCalled()
	}
}

// ExitPruningBufferingMode -
func (sms *StorageManagerStub) ExitPruningBufferingMode() {
	if sms.ExitPruningBufferingModeCalled != nil {
		sms.ExitPruningBufferingModeCalled()
	}
}

// AddDirtyCheckpointHashes -
func (sms *StorageManagerStub) AddDirtyCheckpointHashes(rootHash []byte, hashes data.ModifiedHashes) bool {
	if sms.AddDirtyCheckpointHashesCalled != nil {
		return sms.AddDirtyCheckpointHashesCalled(rootHash, hashes)
	}

	return false
}

// Remove -
func (sms *StorageManagerStub) Remove(hash []byte) error {
	if sms.RemoveCalled != nil {
		return sms.RemoveCalled(hash)
	}

	return nil
}

// GetSnapshotDbBatchDelay -
func (sms *StorageManagerStub) GetSnapshotDbBatchDelay() int {
	return 0
}

// Close -
func (sms *StorageManagerStub) Close() error {
	return nil
}

// IsInterfaceNil -
func (sms *StorageManagerStub) IsInterfaceNil() bool {
	return sms == nil
}
