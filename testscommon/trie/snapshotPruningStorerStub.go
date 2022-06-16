package trie

import (
	"github.com/ElrondNetwork/elrond-go/testscommon"
)

// SnapshotPruningStorerStub -
type SnapshotPruningStorerStub struct {
	*testscommon.MemDbMock
	GetFromOldEpochsWithoutAddingToCacheCalled func(key []byte, epochOffset int) ([]byte, error)
	GetFromEpochWithoutCacheCalled             func(key []byte, epoch uint32) ([]byte, error)
	PutInEpochWithoutCacheCalled               func(key []byte, data []byte, epoch uint32) error
	PutInEpochCalled                           func(key []byte, data []byte, epoch uint32) error
	GetLatestStorageEpochCalled                func() (uint32, error)
	RemoveFromCurrentEpochCalled               func(key []byte) error
}

// GetFromOldEpochsWithoutAddingToCache -
func (spss *SnapshotPruningStorerStub) GetFromOldEpochsWithoutAddingToCache(key []byte, epochOffset int) ([]byte, error) {
	if spss.GetFromOldEpochsWithoutAddingToCacheCalled != nil {
		return spss.GetFromOldEpochsWithoutAddingToCacheCalled(key, epochOffset)
	}

	return nil, nil
}

// PutInEpochWithoutCache -
func (spss *SnapshotPruningStorerStub) PutInEpochWithoutCache(key []byte, data []byte, epoch uint32) error {
	if spss.PutInEpochWithoutCacheCalled != nil {
		return spss.PutInEpochWithoutCacheCalled(key, data, epoch)
	}

	return nil
}

// PutInEpoch -
func (spss *SnapshotPruningStorerStub) PutInEpoch(key []byte, data []byte, epoch uint32) error {
	if spss.PutInEpochCalled != nil {
		return spss.PutInEpochCalled(key, data, epoch)
	}

	return nil
}

// GetLatestStorageEpoch -
func (spss *SnapshotPruningStorerStub) GetLatestStorageEpoch() (uint32, error) {
	if spss.GetLatestStorageEpochCalled != nil {
		return spss.GetLatestStorageEpochCalled()
	}

	return 0, nil
}

// GetFromEpochWithoutCache -
func (spss *SnapshotPruningStorerStub) GetFromEpochWithoutCache(key []byte, epoch uint32) ([]byte, error) {
	if spss.GetFromEpochWithoutCacheCalled != nil {
		return spss.GetFromEpochWithoutCacheCalled(key, epoch)
	}

	return nil, nil
}

// RemoveFromCurrentEpoch -
func (spss *SnapshotPruningStorerStub) RemoveFromCurrentEpoch(key []byte) error {
	if spss.RemoveFromCurrentEpochCalled != nil {
		return spss.RemoveFromCurrentEpochCalled(key)
	}
	return spss.Remove(key)
}
