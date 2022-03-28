package trie

import (
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/storage/memorydb"
)

// SnapshotPruningStorerStub -
type SnapshotPruningStorerStub struct {
	*memorydb.DB
	GetFromOldEpochsWithoutAddingToCacheCalled func(key []byte, priority common.StorageAccessType) ([]byte, error)
	GetFromLastEpochCalled                     func(key []byte, priority common.StorageAccessType) ([]byte, error)
	GetFromCurrentEpochCalled                  func(key []byte, priority common.StorageAccessType) ([]byte, error)
	PutInEpochWithoutCacheCalled               func(key []byte, data []byte, epoch uint32, priority common.StorageAccessType) error
	GetLatestStorageEpochCalled                func() (uint32, error)
	RemoveFromCurrentEpochCalled               func(key []byte, priority common.StorageAccessType) error
}

// GetFromOldEpochsWithoutAddingToCache -
func (spss *SnapshotPruningStorerStub) GetFromOldEpochsWithoutAddingToCache(key []byte, priority common.StorageAccessType) ([]byte, error) {
	if spss.GetFromOldEpochsWithoutAddingToCacheCalled != nil {
		return spss.GetFromOldEpochsWithoutAddingToCacheCalled(key, priority)
	}

	return nil, nil
}

// PutInEpochWithoutCache -
func (spss *SnapshotPruningStorerStub) PutInEpochWithoutCache(key []byte, data []byte, epoch uint32, priority common.StorageAccessType) error {
	if spss.PutInEpochWithoutCacheCalled != nil {
		return spss.PutInEpochWithoutCacheCalled(key, data, epoch, priority)
	}

	return nil
}

// GetFromLastEpoch -
func (spss *SnapshotPruningStorerStub) GetFromLastEpoch(key []byte, priority common.StorageAccessType) ([]byte, error) {
	if spss.GetFromLastEpochCalled != nil {
		return spss.GetFromLastEpochCalled(key, priority)
	}

	return nil, nil
}

// GetFromCurrentEpoch -
func (spss *SnapshotPruningStorerStub) GetFromCurrentEpoch(key []byte, priority common.StorageAccessType) ([]byte, error) {
	if spss.GetFromCurrentEpochCalled != nil {
		return spss.GetFromCurrentEpochCalled(key, priority)
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
func (spss *SnapshotPruningStorerStub) RemoveFromCurrentEpoch(key []byte, priority common.StorageAccessType) error {
	if spss.RemoveFromCurrentEpochCalled != nil {
		return spss.RemoveFromCurrentEpochCalled(key, priority)
	}
	return spss.Remove(key, common.TestPriority)
}
