package testscommon

import (
	"github.com/ElrondNetwork/elrond-go/common"
)

// SnapshotPruningStorerMock -
type SnapshotPruningStorerMock struct {
	*MemDbMock
}

// NewSnapshotPruningStorerMock -
func NewSnapshotPruningStorerMock() *SnapshotPruningStorerMock {
	return &SnapshotPruningStorerMock{NewMemDbMock()}
}

// GetFromOldEpochsWithoutAddingToCache -
func (spsm *SnapshotPruningStorerMock) GetFromOldEpochsWithoutAddingToCache(key []byte, priority common.StorageAccessType) ([]byte, error) {
	return spsm.Get(key, priority)
}

// PutInEpochWithoutCache -
func (spsm *SnapshotPruningStorerMock) PutInEpochWithoutCache(key []byte, data []byte, _ uint32, priority common.StorageAccessType) error {
	return spsm.Put(key, data, priority)
}

// GetFromLastEpoch -
func (spsm *SnapshotPruningStorerMock) GetFromLastEpoch(key []byte, priority common.StorageAccessType) ([]byte, error) {
	return spsm.Get(key, priority)
}

// GetFromCurrentEpoch -
func (spsm *SnapshotPruningStorerMock) GetFromCurrentEpoch(key []byte, priority common.StorageAccessType) ([]byte, error) {
	return spsm.Get(key, priority)
}

// GetLatestStorageEpoch -
func (spsm *SnapshotPruningStorerMock) GetLatestStorageEpoch() (uint32, error) {
	return 0, nil
}

// RemoveFromCurrentEpoch -
func (spsm *SnapshotPruningStorerMock) RemoveFromCurrentEpoch(key []byte, priority common.StorageAccessType) error {
	return spsm.Remove(key, priority)
}
