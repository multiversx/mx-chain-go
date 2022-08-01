package trie

import (
	"github.com/ElrondNetwork/elrond-go/testscommon"
)

// SnapshotPruningStorerStub -
type SnapshotPruningStorerStub struct {
	*testscommon.MemDbMock
	GetFromOldEpochsWithoutAddingToCacheCalled func(key []byte) ([]byte, error)
	GetFromLastEpochCalled                     func(key []byte) ([]byte, error)
	GetFromCurrentEpochCalled                  func(key []byte) ([]byte, error)
	PutInEpochCalled                           func(key []byte, data []byte, epoch uint32) error
	PutInEpochWithoutCacheCalled               func(key []byte, data []byte, epoch uint32) error
	GetLatestStorageEpochCalled                func() (uint32, error)
	RemoveFromCurrentEpochCalled               func(key []byte) error
}

// GetFromOldEpochsWithoutAddingToCache -
func (spss *SnapshotPruningStorerStub) GetFromOldEpochsWithoutAddingToCache(key []byte) ([]byte, error) {
	if spss.GetFromOldEpochsWithoutAddingToCacheCalled != nil {
		return spss.GetFromOldEpochsWithoutAddingToCacheCalled(key)
	}

	return nil, nil
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
