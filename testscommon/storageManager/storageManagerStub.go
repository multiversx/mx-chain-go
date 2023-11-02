package storageManager

import (
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/statistics/disabled"
)

// StorageManagerStub -
type StorageManagerStub struct {
	PutCalled                       func([]byte, []byte) error
	PutInEpochCalled                func([]byte, []byte, uint32) error
	PutInEpochWithoutCacheCalled    func([]byte, []byte, uint32) error
	GetCalled                       func([]byte) ([]byte, error)
	GetFromCurrentEpochCalled       func([]byte) ([]byte, error)
	TakeSnapshotCalled              func(string, []byte, []byte, *common.TrieIteratorChannels, chan []byte, common.SnapshotStatisticsHandler, uint32)
	GetDbThatContainsHashCalled     func([]byte) common.BaseStorer
	IsPruningEnabledCalled          func() bool
	IsPruningBlockedCalled          func() bool
	EnterPruningBufferingModeCalled func()
	ExitPruningBufferingModeCalled  func()
	RemoveFromCurrentEpochCalled    func([]byte) error
	RemoveCalled                    func([]byte) error
	IsInterfaceNilCalled            func() bool
	SetEpochForPutOperationCalled   func(uint32)
	ShouldTakeSnapshotCalled        func() bool
	GetLatestStorageEpochCalled     func() (uint32, error)
	IsClosedCalled                  func() bool
	GetBaseTrieStorageManagerCalled func() common.StorageManager
	GetIdentifierCalled             func() string
	CloseCalled                     func() error
	RemoveFromAllActiveEpochsCalled func(hash []byte) error
	GetStateStatsHandlerCalled      func() common.StateStatisticsHandler
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

// PutInEpochWithoutCache -
func (sms *StorageManagerStub) PutInEpochWithoutCache(key []byte, val []byte, epoch uint32) error {
	if sms.PutInEpochWithoutCacheCalled != nil {
		return sms.PutInEpochWithoutCacheCalled(key, val, epoch)
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
func (sms *StorageManagerStub) TakeSnapshot(
	address string,
	rootHash []byte,
	mainTrieRootHash []byte,
	iteratorChannels *common.TrieIteratorChannels,
	missingNodesChan chan []byte,
	stats common.SnapshotStatisticsHandler,
	epoch uint32,
) {
	if sms.TakeSnapshotCalled != nil {
		sms.TakeSnapshotCalled(address, rootHash, mainTrieRootHash, iteratorChannels, missingNodesChan, stats, epoch)
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

// RemoveFromCurrentEpoch -
func (sms *StorageManagerStub) RemoveFromCurrentEpoch(hash []byte) error {
	if sms.RemoveFromCurrentEpochCalled != nil {
		return sms.RemoveFromCurrentEpochCalled(hash)
	}
	return nil
}

// Remove -
func (sms *StorageManagerStub) Remove(hash []byte) error {
	if sms.RemoveCalled != nil {
		return sms.RemoveCalled(hash)
	}

	return nil
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
	if sms.CloseCalled != nil {
		return sms.CloseCalled()
	}
	return nil
}

// IsClosed -
func (sms *StorageManagerStub) IsClosed() bool {
	if sms.IsClosedCalled != nil {
		return sms.IsClosedCalled()
	}

	return false
}

// GetBaseTrieStorageManager -
func (sms *StorageManagerStub) GetBaseTrieStorageManager() common.StorageManager {
	if sms.GetBaseTrieStorageManagerCalled != nil {
		return sms.GetBaseTrieStorageManagerCalled()
	}

	return nil
}

// RemoveFromAllActiveEpochs -
func (sms *StorageManagerStub) RemoveFromAllActiveEpochs(hash []byte) error {
	if sms.RemoveFromAllActiveEpochsCalled != nil {
		return sms.RemoveFromAllActiveEpochsCalled(hash)
	}

	return nil
}

// GetIdentifier -
func (sms *StorageManagerStub) GetIdentifier() string {
	if sms.GetIdentifierCalled != nil {
		return sms.GetIdentifierCalled()
	}

	return ""
}

// GetStateStatsHandler -
func (sms *StorageManagerStub) GetStateStatsHandler() common.StateStatisticsHandler {
	if sms.GetStateStatsHandlerCalled != nil {
		return sms.GetStateStatsHandlerCalled()
	}

	return disabled.NewStateStatistics()
}

// IsInterfaceNil -
func (sms *StorageManagerStub) IsInterfaceNil() bool {
	return sms == nil
}
