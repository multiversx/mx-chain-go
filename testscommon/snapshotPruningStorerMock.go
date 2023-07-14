package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/core"
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
func (spsm *SnapshotPruningStorerMock) GetFromOldEpochsWithoutAddingToCache(key []byte) ([]byte, core.OptionalUint32, error) {
	val, err := spsm.Get(key)

	return val, core.OptionalUint32{}, err
}

// PutInEpoch -
func (spsm *SnapshotPruningStorerMock) PutInEpoch(key []byte, data []byte, _ uint32) error {
	return spsm.Put(key, data)
}

// PutInEpochWithoutCache -
func (spsm *SnapshotPruningStorerMock) PutInEpochWithoutCache(key []byte, data []byte, _ uint32) error {
	return spsm.Put(key, data)
}

// GetFromLastEpoch -
func (spsm *SnapshotPruningStorerMock) GetFromLastEpoch(key []byte) ([]byte, error) {
	return spsm.Get(key)
}

// GetFromCurrentEpoch -
func (spsm *SnapshotPruningStorerMock) GetFromCurrentEpoch(key []byte) ([]byte, error) {
	return spsm.Get(key)
}

// GetFromEpoch -
func (spsm *SnapshotPruningStorerMock) GetFromEpoch(key []byte, _ uint32) ([]byte, error) {
	return spsm.Get(key)
}

// GetLatestStorageEpoch -
func (spsm *SnapshotPruningStorerMock) GetLatestStorageEpoch() (uint32, error) {
	return 0, nil
}

// RemoveFromAllActiveEpochs -
func (spsm *SnapshotPruningStorerMock) RemoveFromAllActiveEpochs(_ []byte) error {
	return nil
}

// RemoveFromCurrentEpoch -
func (spsm *SnapshotPruningStorerMock) RemoveFromCurrentEpoch(key []byte) error {
	return spsm.Remove(key)
}

// GetWithStats -
func (spsm *SnapshotPruningStorerMock) GetWithStats(key []byte) ([]byte, bool, error) {
	v, err := spsm.Get(key)
	return v, false, err
}
