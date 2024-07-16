package trieChangesBatch

import (
	"sort"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
)

type trieChangesBatch struct {
	insertedData map[string]core.TrieData
	deletedKeys  map[string]struct{}

	mutex sync.RWMutex
}

// NewTrieChangesBatch creates a new instance of trieChangesBatch
func NewTrieChangesBatch() *trieChangesBatch {
	return &trieChangesBatch{
		insertedData: make(map[string]core.TrieData),
		deletedKeys:  make(map[string]struct{}),
	}
}

// Add adds a new key and data to the batch
func (t *trieChangesBatch) Add(key []byte, data core.TrieData) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	_, ok := t.deletedKeys[string(key)]
	if ok {
		delete(t.deletedKeys, string(key))
	}

	t.insertedData[string(key)] = data
}

// MarkForRemoval marks the key for removal
func (t *trieChangesBatch) MarkForRemoval(key []byte) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	_, ok := t.insertedData[string(key)]
	if ok {
		delete(t.insertedData, string(key))
	}

	t.deletedKeys[string(key)] = struct{}{}
}

// Get returns the data for the given key
func (t *trieChangesBatch) Get(key []byte) ([]byte, bool) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	data, ok := t.insertedData[string(key)]
	if ok {
		return data.Value, true
	}

	_, ok = t.deletedKeys[string(key)]
	if ok {
		return nil, true
	}

	return nil, false
}

// GetSortedDataForInsertion returns the data sorted for insertion
func (t *trieChangesBatch) GetSortedDataForInsertion() ([]string, map[string]core.TrieData) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	keys := make([]string, 0, len(t.insertedData))
	for k := range t.insertedData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys, t.insertedData
}

// GetSortedDataForRemoval returns the data sorted for removal
func (t *trieChangesBatch) GetSortedDataForRemoval() []string {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	keys := make([]string, 0, len(t.deletedKeys))
	for k := range t.deletedKeys {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

// IsInterfaceNil returns true if there is no value under the interface
func (t *trieChangesBatch) IsInterfaceNil() bool {
	return t == nil
}
