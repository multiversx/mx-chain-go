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

// PutWithoutCache -
func (spsm *SnapshotPruningStorerMock) PutWithoutCache(key, data []byte) error {
	return spsm.Put(key, data)
}

// GetFromLastEpoch -
func (spsm *SnapshotPruningStorerMock) GetFromLastEpoch(key []byte) ([]byte, error) {
	return spsm.Get(key)
}
