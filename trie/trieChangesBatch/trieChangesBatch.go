package trieChangesBatch

import (
	"bytes"
	"sort"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/trie/keyBuilder"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("trieBatch")

type trieChangesBatch struct {
	insertedData map[string]core.TrieData
	deletedKeys  map[string]core.TrieData

	identifier string
	mutex      sync.RWMutex
}

// NewTrieChangesBatch creates a new instance of trieChangesBatch
func NewTrieChangesBatch(identifier string) *trieChangesBatch {
	return &trieChangesBatch{
		insertedData: make(map[string]core.TrieData),
		deletedKeys:  make(map[string]core.TrieData),
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

	t.deletedKeys[string(key)] = core.TrieData{Key: key}
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

	if len(t.insertedData) == 0 {
		return nil
	}

	log.Trace("sorted data for insertion", "identifier", t.identifier, "num insertions", len(t.insertedData))

	return getSortedData(t.insertedData)
}

// GetSortedDataForRemoval returns the data sorted for removal
func (t *trieChangesBatch) GetSortedDataForRemoval() []core.TrieData {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	if len(t.deletedKeys) == 0 {
		return nil
	}

	log.Trace("sorted data for removal", "identifier", t.identifier, "num deletes", len(t.deletedKeys))

	return getSortedData(t.deletedKeys)
}

// IsInterfaceNil returns true if there is no value under the interface
func (t *trieChangesBatch) IsInterfaceNil() bool {
	return t == nil
}

func getSortedData(dataMap map[string]core.TrieData) []core.TrieData {
	data := make([]core.TrieData, 0, len(dataMap))
	for _, trieData := range dataMap {
		trieData.Key = keyBuilder.KeyBytesToHex(trieData.Key)
		data = append(data, trieData)
	}

	sort.Slice(data, func(i, j int) bool {
		return bytes.Compare(data[i].Key, data[j].Key) < 0
	})

	return data
}
