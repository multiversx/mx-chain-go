package trie

import "github.com/ElrondNetwork/elrond-go/storage/memorydb"

// SnapshotPruningStorerStub -
type SnapshotPruningStorerStub struct {
	memorydb.DB
	GetFromOldEpochsWithoutAddingToCacheCalled func(key []byte) ([]byte, error)
	GetFromLastEpochCalled             func(key []byte) ([]byte, error)
	PutInEpochWithoutCacheCalled       func(key []byte, data []byte, epoch uint32) error
	GetLatestStorageEpochCalled        func() (uint32, error)
}

// GetFromOldEpochsWithoutAddingToCache -
func (spss *SnapshotPruningStorerStub) GetFromOldEpochsWithoutAddingToCache(key []byte) ([]byte, error) {
	if spss.GetFromOldEpochsWithoutAddingToCacheCalled != nil {
		return spss.GetFromOldEpochsWithoutAddingToCacheCalled(key)
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

// GetFromLastEpoch -
func (spss *SnapshotPruningStorerStub) GetFromLastEpoch(key []byte) ([]byte, error) {
	if spss.GetFromLastEpochCalled != nil {
		return spss.GetFromLastEpochCalled(key)
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
