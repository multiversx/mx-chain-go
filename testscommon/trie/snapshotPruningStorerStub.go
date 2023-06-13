package trie

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/testscommon"
)

// SnapshotPruningStorerStub -
type SnapshotPruningStorerStub struct {
	*testscommon.MemDbMock
	GetFromOldEpochsWithoutAddingToCacheCalled func(key []byte) ([]byte, core.OptionalUint32, error)
	GetFromLastEpochCalled                     func(key []byte) ([]byte, error)
	GetFromCurrentEpochCalled                  func(key []byte) ([]byte, error)
	GetFromEpochCalled                         func(key []byte, epoch uint32) ([]byte, error)
	PutInEpochCalled                           func(key []byte, data []byte, epoch uint32) error
	PutInEpochWithoutCacheCalled               func(key []byte, data []byte, epoch uint32) error
	GetLatestStorageEpochCalled                func() (uint32, error)
	RemoveFromCurrentEpochCalled               func(key []byte) error
	CloseCalled                                func() error
	RemoveFromAllActiveEpochsCalled            func(key []byte) error
}

// GetFromOldEpochsWithoutAddingToCache -
func (spss *SnapshotPruningStorerStub) GetFromOldEpochsWithoutAddingToCache(key []byte) ([]byte, core.OptionalUint32, error) {
	if spss.GetFromOldEpochsWithoutAddingToCacheCalled != nil {
		return spss.GetFromOldEpochsWithoutAddingToCacheCalled(key)
	}

	return nil, core.OptionalUint32{}, nil
}

// PutInEpoch -
func (spss *SnapshotPruningStorerStub) PutInEpoch(key []byte, data []byte, epoch uint32) error {
	if spss.PutInEpochCalled != nil {
		return spss.PutInEpochCalled(key, data, epoch)
	}

	return nil
}

// PutInEpochWithoutCache -
func (spss *SnapshotPruningStorerStub) PutInEpochWithoutCache(key []byte, data []byte, epoch uint32) error {
	if spss.PutInEpochWithoutCacheCalled != nil {
		return spss.PutInEpochWithoutCacheCalled(key, data, epoch)
	}

	return nil
}

// GetFromLastEpoch -
func (spss *SnapshotPruningStorerStub) GetFromLastEpoch(key []byte) ([]byte, error) {
	if spss.GetFromLastEpochCalled != nil {
		return spss.GetFromLastEpochCalled(key)
	}

	return nil, nil
}

// GetFromCurrentEpoch -
func (spss *SnapshotPruningStorerStub) GetFromCurrentEpoch(key []byte) ([]byte, error) {
	if spss.GetFromCurrentEpochCalled != nil {
		return spss.GetFromCurrentEpochCalled(key)
	}

	return nil, nil
}

// GetFromEpoch -
func (spss *SnapshotPruningStorerStub) GetFromEpoch(key []byte, epoch uint32) ([]byte, error) {
	if spss.GetFromEpochCalled != nil {
		return spss.GetFromEpochCalled(key, epoch)
	}

	return nil, nil
}

// GetLatestStorageEpoch -
func (spss *SnapshotPruningStorerStub) GetLatestStorageEpoch() (uint32, error) {
	if spss.GetLatestStorageEpochCalled != nil {
		return spss.GetLatestStorageEpochCalled()
	}

	return 0, nil
}

// RemoveFromCurrentEpoch -
func (spss *SnapshotPruningStorerStub) RemoveFromCurrentEpoch(key []byte) error {
	if spss.RemoveFromCurrentEpochCalled != nil {
		return spss.RemoveFromCurrentEpochCalled(key)
	}
	return spss.Remove(key)
}

// Close -
func (spss *SnapshotPruningStorerStub) Close() error {
	if spss.CloseCalled != nil {
		return spss.CloseCalled()
	}
	return nil
}

// RemoveFromAllActiveEpochs -
func (spss *SnapshotPruningStorerStub) RemoveFromAllActiveEpochs(key []byte) error {
	if spss.RemoveFromAllActiveEpochsCalled != nil {
		return spss.RemoveFromAllActiveEpochsCalled(key)
	}

	return spss.Remove(key)
}
