package trieBatchManager

import (
	"errors"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/trie/trieChangesBatch"
)

// ErrTrieUpdateInProgress signals that a trie update is in progress
var ErrTrieUpdateInProgress = errors.New("trie update is in progress")

type trieBatchManager struct {
	currentBatch common.TrieBatcher
	tempBatch    common.TrieBatcher

	isUpdateIsProgress bool
	mutex              sync.RWMutex
}

// NewTrieBatchManager creates a new instance of trieBatchManager
func NewTrieBatchManager() *trieBatchManager {
	return &trieBatchManager{
		currentBatch:       trieChangesBatch.NewTrieChangesBatch(),
		tempBatch:          nil,
		isUpdateIsProgress: false,
	}
}

// MarkTrieUpdateInProgress marks the trie update in progress and returns the current batch
func (t *trieBatchManager) MarkTrieUpdateInProgress() (common.TrieBatcher, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.isUpdateIsProgress {
		return nil, ErrTrieUpdateInProgress
	}

	t.isUpdateIsProgress = true
	t.tempBatch = t.currentBatch
	t.currentBatch = trieChangesBatch.NewTrieChangesBatch()

	return t.tempBatch, nil
}

// MarkTrieUpdateCompleted marks the trie update as completed
func (t *trieBatchManager) MarkTrieUpdateCompleted() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.isUpdateIsProgress = false
	t.tempBatch = nil
}

// Add adds a new key and data to the current batch
func (t *trieBatchManager) Add(key []byte, data core.TrieData) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	t.currentBatch.Add(key, data)
}

// Get returns the data for the given key checking first the current batch and then the temp batch if the trie update is in progress
func (t *trieBatchManager) Get(key []byte) ([]byte, bool) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	val, isPresent := t.currentBatch.Get(key)
	if isPresent {
		return val, true
	}

	if t.isUpdateIsProgress && !check.IfNil(t.tempBatch) {
		val, isPresent = t.tempBatch.Get(key)
		if isPresent {
			return val, true
		}
	}

	return nil, false
}

// MarkForRemoval marks the key for removal in the current batch
func (t *trieBatchManager) MarkForRemoval(key []byte) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	t.currentBatch.MarkForRemoval(key)
}

// IsInterfaceNil returns true if there is no value under the interface
func (t *trieBatchManager) IsInterfaceNil() bool {
	return t == nil
}
