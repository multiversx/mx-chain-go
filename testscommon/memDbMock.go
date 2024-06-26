package testscommon

import (
	"encoding/base64"
	"errors"
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/statistics/disabled"
)

// MemDbMock represents the memory database storage. It holds a map of key value pairs
// and a mutex to handle concurrent accesses to the map
type MemDbMock struct {
	db                         map[string][]byte
	mutx                       sync.RWMutex
	PutCalled                  func(key, val []byte) error
	GetCalled                  func(key []byte) ([]byte, error)
	GetIdentifierCalled        func() string
	GetStateStatsHandlerCalled func() common.StateStatisticsHandler
}

// NewMemDbMock creates a new memorydb object
func NewMemDbMock() *MemDbMock {
	return &MemDbMock{
		db:   make(map[string][]byte),
		mutx: sync.RWMutex{},
	}
}

// Put adds the value to the (key, val) storage medium
func (s *MemDbMock) Put(key, val []byte) error {
	s.mutx.Lock()
	defer s.mutx.Unlock()

	s.db[string(key)] = val

	if s.PutCalled != nil {
		return s.PutCalled(key, val)
	}

	return nil
}

// Get gets the value associated to the key, or reports an error
func (s *MemDbMock) Get(key []byte) ([]byte, error) {
	s.mutx.RLock()
	defer s.mutx.RUnlock()

	val, ok := s.db[string(key)]

	if s.GetCalled != nil {
		return s.GetCalled(key)
	}

	if !ok {
		return nil, fmt.Errorf("key: %s not found", base64.StdEncoding.EncodeToString(key))
	}

	return val, nil
}

// Has returns true if the given key is present in the persistence medium, false otherwise
func (s *MemDbMock) Has(key []byte) error {
	s.mutx.RLock()
	defer s.mutx.RUnlock()

	_, ok := s.db[string(key)]
	if !ok {
		return errors.New("key not present")
	}

	return nil
}

// Close closes the files/resources associated to the storage medium
func (s *MemDbMock) Close() error {
	// nothing to do
	return nil
}

// Remove removes the data associated to the given key
func (s *MemDbMock) Remove(key []byte) error {
	s.mutx.Lock()
	defer s.mutx.Unlock()

	delete(s.db, string(key))

	return nil
}

// Destroy removes the storage medium stored data
func (s *MemDbMock) Destroy() error {
	s.mutx.Lock()
	defer s.mutx.Unlock()

	s.db = make(map[string][]byte)

	return nil
}

// DestroyClosed removes the already closed storage medium stored data
func (s *MemDbMock) DestroyClosed() error {
	return nil
}

// RangeKeys will iterate over all contained (key, value) pairs calling the handler for each pair
func (s *MemDbMock) RangeKeys(handler func(key []byte, value []byte) bool) {
	if handler == nil {
		return
	}

	s.mutx.RLock()
	defer s.mutx.RUnlock()

	for k, v := range s.db {
		shouldContinue := handler([]byte(k), v)
		if !shouldContinue {
			return
		}
	}
}

// GetIdentifier returns the identifier of the storage medium
func (s *MemDbMock) GetIdentifier() string {
	if s.GetIdentifierCalled != nil {
		return s.GetIdentifierCalled()
	}

	return ""
}

// GetStateStatsHandler -
func (s *MemDbMock) GetStateStatsHandler() common.StateStatisticsHandler {
	if s.GetStateStatsHandlerCalled != nil {
		return s.GetStateStatsHandlerCalled()
	}

	return disabled.NewStateStatistics()
}

// IsInterfaceNil returns true if there is no value under the interface
func (s *MemDbMock) IsInterfaceNil() bool {
	return s == nil
}
