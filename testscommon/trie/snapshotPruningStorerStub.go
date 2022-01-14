package trie

import "github.com/ElrondNetwork/elrond-go/storage/memorydb"

// SnapshotPruningStorerStub -
type SnapshotPruningStorerStub struct {
	memorydb.DB
	GetFromOldEpochsWithoutAddingToCacheCalled func(key []byte) ([]byte, error)
	GetFromLastEpochCalled                     func(key []byte) ([]byte, error)
	PutWithoutCacheCalled                      func(key, data []byte) error
}

// GetFromOldEpochsWithoutAddingToCache -
func (spss *SnapshotPruningStorerStub) GetFromOldEpochsWithoutAddingToCache(key []byte) ([]byte, error) {
	if spss.GetFromOldEpochsWithoutAddingToCacheCalled != nil {
		return spss.GetFromOldEpochsWithoutAddingToCacheCalled(key)
	}

	return nil, nil
}

// PutWithoutCache -
func (spss *SnapshotPruningStorerStub) PutWithoutCache(key, data []byte) error {
	if spss.PutWithoutCacheCalled != nil {
		return spss.PutWithoutCacheCalled(key, data)
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
