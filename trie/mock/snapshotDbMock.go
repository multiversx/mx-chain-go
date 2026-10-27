package mock

import (
	"github.com/multiversx/mx-chain-go/testscommon"
)

// SnapshotDbMock represents the memory database storage. It holds a map of key value pairs
type SnapshotDbMock struct {
	testscommon.MemDbMock
}

// NewSnapshotDbMock creates a new memorydb object
func NewSnapshotDbMock() *SnapshotDbMock {
	return &SnapshotDbMock{
		MemDbMock: *testscommon.NewMemDbMock(),
	}
}

// GetWithoutAddingToCache gets the value associated to the key from old epochs, or reports an error
func (s *SnapshotDbMock) GetWithoutAddingToCache(key []byte, _ uint32) ([]byte, uint32, error) {
	val, err := s.MemDbMock.Get(key)
	return val, 0, err
}

// GetFromLastEpoch gets the value associated to the key from the last epoch, or reports an error
func (s *SnapshotDbMock) GetFromLastEpoch(key []byte) ([]byte, error) {
	return s.MemDbMock.Get(key)
}

// PutInEpochWithoutCache adds the given (key, data) in the current epoch without adding it to cache
func (s *SnapshotDbMock) PutInEpochWithoutCache(key, data []byte) error {
	return s.MemDbMock.Put(key, data)
}

// GetIdentifier returns the identifier of the storage medium
func (s *SnapshotDbMock) GetIdentifier() string {
	return s.MemDbMock.GetIdentifier()
}
