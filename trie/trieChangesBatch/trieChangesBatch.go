package trieChangesBatch

import (
	"bytes"
	"sort"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("trieBatch")

type trieChangesBatch struct {
	insertedData map[string]core.TrieData
	deletedKeys  map[string]struct{}

	identifier string
	mutex      sync.RWMutex
}

// NewTrieChangesBatch creates a new instance of trieChangesBatch
func NewTrieChangesBatch(identifier string) *trieChangesBatch {
	return &trieChangesBatch{
		insertedData: make(map[string]core.TrieData),
		deletedKeys:  make(map[string]struct{}),
		identifier:   identifier,
	}
}

// Add adds a new key and data to the batch
func (t *trieChangesBatch) Add(data core.TrieData) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	_, ok := t.deletedKeys[string(data.Key)]
	if ok {
		delete(t.deletedKeys, string(data.Key))
	}

	t.insertedData[string(data.Key)] = data
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
func (t *trieChangesBatch) GetSortedDataForInsertion() []core.TrieData {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	data := make([]core.TrieData, 0, len(t.insertedData))
	for k := range t.insertedData {
		data = append(data, t.insertedData[k])
	}

	log.Trace("sorted data for insertion", "identifier", t.identifier, "num insertions", len(data))

	return getSortedData(data)
}

// GetSortedDataForRemoval returns the data sorted for removal
func (t *trieChangesBatch) GetSortedDataForRemoval() []core.TrieData {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	data := make([]core.TrieData, 0, len(t.deletedKeys))
	for k := range t.deletedKeys {
		data = append(data, core.TrieData{Key: []byte(k)})
	}

	log.Trace("sorted data for removal", "identifier", t.identifier, "num deletes", len(data))

	return getSortedData(data)
}

// IsInterfaceNil returns true if there is no value under the interface
func (t *trieChangesBatch) IsInterfaceNil() bool {
	return t == nil
}

func getSortedData(data []core.TrieData) []core.TrieData {
	sort.Slice(data, func(i, j int) bool {
		return bytes.Compare(data[i].Key, data[j].Key) < 0
	})

	return data
}
