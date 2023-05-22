package state

import (
	"sync"

	"github.com/multiversx/mx-chain-go/common"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type dataTriesHolder struct {
	tries map[string]common.Trie
	mutex sync.RWMutex
}

// NewDataTriesHolder creates a new instance of dataTriesHolder
func NewDataTriesHolder() *dataTriesHolder {
	return &dataTriesHolder{
		tries: make(map[string]common.Trie),
	}
}

// Put adds a trie pointer to the tries map
func (dth *dataTriesHolder) Put(key []byte, tr common.Trie) {
	log.Trace("put trie in data tries holder", "key", key)

	dth.mutex.Lock()
	dth.tries[string(key)] = tr
	dth.mutex.Unlock()
}

// Replace changes a trie pointer to the tries map
func (dth *dataTriesHolder) Replace(key []byte, tr common.Trie) {
	dth.Put(key, tr)
}

// Get returns the trie pointer that is stored in the map at the given key
func (dth *dataTriesHolder) Get(key []byte) common.Trie {
	dth.mutex.Lock()
	defer dth.mutex.Unlock()

	return dth.tries[string(key)]
}

// GetAll returns all trie pointers from the map
func (dth *dataTriesHolder) GetAll() []common.Trie {
	dth.mutex.Lock()
	defer dth.mutex.Unlock()

	tries := make([]common.Trie, 0)
	for _, trie := range dth.tries {
		tries = append(tries, trie)
	}

	return tries
}

// GetAllTries returns the tries with key value map
func (dth *dataTriesHolder) GetAllTries() map[string]common.Trie {
	dth.mutex.Lock()
	defer dth.mutex.Unlock()

	copyTries := make(map[string]common.Trie, len(dth.tries))
	for key, trie := range dth.tries {
		copyTries[key] = trie
	}

	return copyTries
}

// Reset clears the tries map
func (dth *dataTriesHolder) Reset() {
	dth.mutex.Lock()

	if log.GetLevel() == logger.LogTrace {
		for key := range dth.tries {
			log.Trace("reset data tries holder", "key", key)
		}
	}

	dth.tries = make(map[string]common.Trie)
	dth.mutex.Unlock()
}

// IsInterfaceNil returns true if underlying object is nil
func (dth *dataTriesHolder) IsInterfaceNil() bool {
	return dth == nil
}
