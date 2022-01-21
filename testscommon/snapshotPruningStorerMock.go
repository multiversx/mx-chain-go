package testscommon

// SnapshotPruningStorerMock -
type SnapshotPruningStorerMock struct {
	*MemDbMock
}

// NewSnapshotPruningStorerMock -
func NewSnapshotPruningStorerMock() *SnapshotPruningStorerMock {
	return &SnapshotPruningStorerMock{NewMemDbMock()}
}

// GetFromOldEpochsWithoutAddingToCache -
func (spsm *SnapshotPruningStorerMock) GetFromOldEpochsWithoutAddingToCache(key []byte) ([]byte, error) {
	return spsm.Get(key)
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

// GetLatestStorageEpoch -
func (spsm *SnapshotPruningStorerMock) GetLatestStorageEpoch() (uint32, error) {
	return 0, nil
}
