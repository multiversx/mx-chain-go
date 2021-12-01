package trie

import "github.com/ElrondNetwork/elrond-go/storage/memorydb"

// SnapshotPruningStorerStub -
type SnapshotPruningStorerStub struct {
	memorydb.DB
	GetFromOldEpochsWithoutCacheCalled func(key []byte) ([]byte, error)
	PutWithoutCacheCalled              func(key, data []byte) error
}

// GetFromOldEpochsWithoutCache -
func (spss *SnapshotPruningStorerStub) GetFromOldEpochsWithoutCache(key []byte) ([]byte, error) {
	if spss.GetFromOldEpochsWithoutCacheCalled != nil {
		return spss.GetFromOldEpochsWithoutCacheCalled(key)
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
