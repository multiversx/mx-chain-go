package testscommon

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/common"
)

// StorageManagerStub -
type StorageManagerStub struct {
	PutCalled                       func([]byte, []byte) error
	GetCalled                       func([]byte) ([]byte, error)
	TakeSnapshotCalled              func([]byte, bool, chan core.KeyValueHolder)
	SetCheckpointCalled             func([]byte, chan core.KeyValueHolder)
	GetDbThatContainsHashCalled     func([]byte) common.DBWriteCacher
	IsPruningEnabledCalled          func() bool
	IsPruningBlockedCalled          func() bool
	EnterPruningBufferingModeCalled func()
	ExitPruningBufferingModeCalled  func()
	AddDirtyCheckpointHashesCalled  func([]byte, common.ModifiedHashes) bool
	RemoveCalled                    func([]byte) error
	IsInterfaceNilCalled            func() bool
	SetEpochForPutOperationCalled   func(uint32)
}

// Put -
func (sms *StorageManagerStub) Put(key []byte, val []byte) error {
	if sms.PutCalled != nil {
		return sms.PutCalled(key, val)
	}

	return nil
}

// Get -
func (sms *StorageManagerStub) Get(key []byte) ([]byte, error) {
	if sms.GetCalled != nil {
		return sms.GetCalled(key)
	}

	return nil, nil
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
func (sms *StorageManagerStub) AddDirtyCheckpointHashes(rootHash []byte, hashes common.ModifiedHashes) bool {
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

// SetEpochForPutOperation -
func (sms *StorageManagerStub) SetEpochForPutOperation(epoch uint32) {
	if sms.SetEpochForPutOperationCalled != nil {
		sms.SetEpochForPutOperationCalled(epoch)
	}
}

// Close -
func (sms *StorageManagerStub) Close() error {
	return nil
}

// IsInterfaceNil -
func (sms *StorageManagerStub) IsInterfaceNil() bool {
	return sms == nil
}
