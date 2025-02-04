package trieBatchManager

import (
	"errors"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/trie/trieChangesBatch"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("trieBatch")

// ErrTrieUpdateInProgress signals that a trie update is in progress
var ErrTrieUpdateInProgress = errors.New("trie update is in progress")

type trieBatchManager struct {
	currentBatch common.TrieBatcher
	tempBatch    common.TrieBatcher

	isUpdateInProgress bool
	identifier         string
	mutex              sync.RWMutex
}

// NewTrieBatchManager creates a new instance of trieBatchManager
func NewTrieBatchManager(identifier string) *trieBatchManager {
	return &trieBatchManager{
		currentBatch:       trieChangesBatch.NewTrieChangesBatch(identifier),
		tempBatch:          nil,
		isUpdateInProgress: false,
		identifier:         identifier,
	}
}

// MarkTrieUpdateInProgress marks the trie update in progress and returns the current batch
func (t *trieBatchManager) MarkTrieUpdateInProgress() (common.TrieBatcher, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.isUpdateInProgress {
		return nil, ErrTrieUpdateInProgress
	}

	log.Trace("marking trie update in progress", "identifier", t.identifier)

	t.isUpdateInProgress = true
	t.tempBatch = t.currentBatch
	t.currentBatch = trieChangesBatch.NewTrieChangesBatch(t.identifier)

	return t.tempBatch, nil
}

// MarkTrieUpdateCompleted marks the trie update as completed
func (t *trieBatchManager) MarkTrieUpdateCompleted() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	log.Trace("marking trie update completed", "identifier", t.identifier)

	t.isUpdateInProgress = false
	t.tempBatch = nil
}

// Add adds a new key and data to the current batch
func (t *trieBatchManager) Add(data core.TrieData) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	log.Trace("adding data to the current batch", "identifier", t.identifier, "key", data.Key, "value", data.Value)

	t.currentBatch.Add(data)
}

// Get returns the data for the given key checking first the current batch and then the temp batch if the trie update is in progress
func (t *trieBatchManager) Get(key []byte) ([]byte, bool) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	val, isPresent := t.currentBatch.Get(key)
	if isPresent {
		log.Trace("found data in the current batch", "identifier", t.identifier, "key", key, "value", val)
		return val, true
	}

	if t.isUpdateInProgress && !check.IfNil(t.tempBatch) {
		val, isPresent = t.tempBatch.Get(key)
		if isPresent {
			log.Trace("found data in the temp batch", "identifier", t.identifier, "key", key, "value", val)
			return val, true
		}
	}

	log.Trace("data not found in any batch", "identifier", t.identifier, "key", key)
	return nil, false
}

// MarkForRemoval marks the key for removal in the current batch
func (t *trieBatchManager) MarkForRemoval(key []byte) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	log.Trace("marking key for removal in the current batch", "identifier", t.identifier, "key", key)

	t.currentBatch.MarkForRemoval(key)
}

// IsInterfaceNil returns true if there is no value under the interface
func (t *trieBatchManager) IsInterfaceNil() bool {
	return t == nil
}
