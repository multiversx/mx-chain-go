package testscommon

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/common"
)

// StorageManagerStub -
type StorageManagerStub struct {
	PutCalled                       func([]byte, []byte) error
	PutInEpochCalled                func([]byte, []byte, uint32) error
	GetCalled                       func([]byte) ([]byte, error)
	GetFromCurrentEpochCalled       func([]byte) ([]byte, error)
	TakeSnapshotCalled              func([]byte, []byte, chan core.KeyValueHolder, common.SnapshotStatisticsHandler, uint32)
	SetCheckpointCalled             func([]byte, []byte, chan core.KeyValueHolder, common.SnapshotStatisticsHandler)
	GetDbThatContainsHashCalled     func([]byte) common.DBWriteCacher
	IsPruningEnabledCalled          func() bool
	IsPruningBlockedCalled          func() bool
	EnterPruningBufferingModeCalled func()
	ExitPruningBufferingModeCalled  func()
	AddDirtyCheckpointHashesCalled  func([]byte, common.ModifiedHashes) bool
	RemoveCalled                    func([]byte) error
	IsInterfaceNilCalled            func() bool
	SetEpochForPutOperationCalled   func(uint32)
	ShouldTakeSnapshotCalled        func() bool
	GetLatestStorageEpochCalled     func() (uint32, error)
}

// Put -
func (sms *StorageManagerStub) Put(key []byte, val []byte) error {
	if sms.PutCalled != nil {
		return sms.PutCalled(key, val)
	}

	return nil
}

// PutInEpoch -
func (sms *StorageManagerStub) PutInEpoch(key []byte, val []byte, epoch uint32) error {
	if sms.PutInEpochCalled != nil {
		return sms.PutInEpochCalled(key, val, epoch)
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

// GetFromCurrentEpoch -
func (sms *StorageManagerStub) GetFromCurrentEpoch(key []byte) ([]byte, error) {
	if sms.GetFromCurrentEpochCalled != nil {
		return sms.GetFromCurrentEpochCalled(key)
	}

	return nil, nil
}

// TakeSnapshot -
func (sms *StorageManagerStub) TakeSnapshot(rootHash []byte, mainTrieRootHash []byte, leavesChan chan core.KeyValueHolder, stats common.SnapshotStatisticsHandler, epoch uint32) {
	if sms.TakeSnapshotCalled != nil {
		sms.TakeSnapshotCalled(rootHash, mainTrieRootHash, leavesChan, stats, epoch)
	}
}

// SetCheckpoint -
func (sms *StorageManagerStub) SetCheckpoint(rootHash []byte, mainTrieRootHash []byte, leavesChan chan core.KeyValueHolder, stats common.SnapshotStatisticsHandler) {
	if sms.SetCheckpointCalled != nil {
		sms.SetCheckpointCalled(rootHash, mainTrieRootHash, leavesChan, stats)
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

// ShouldTakeSnapshot -
func (sms *StorageManagerStub) ShouldTakeSnapshot() bool {
	if sms.ShouldTakeSnapshotCalled != nil {
		return sms.ShouldTakeSnapshotCalled()
	}

	return true
}

// GetLatestStorageEpoch -
func (sms *StorageManagerStub) GetLatestStorageEpoch() (uint32, error) {
	if sms.GetLatestStorageEpochCalled != nil {
		return sms.GetLatestStorageEpochCalled()
	}

	return 0, nil
}

// Close -
func (sms *StorageManagerStub) Close() error {
	return nil
}

// IsInterfaceNil -
func (sms *StorageManagerStub) IsInterfaceNil() bool {
	return sms == nil
}
