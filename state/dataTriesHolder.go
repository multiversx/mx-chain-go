package state

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/state/temporary"
)

type dataTriesHolder struct {
	tries map[string]temporary.Trie
	mutex sync.RWMutex
}

// NewDataTriesHolder creates a new instance of dataTriesHolder
func NewDataTriesHolder() *dataTriesHolder {
	return &dataTriesHolder{
		tries: make(map[string]temporary.Trie),
	}
}

// Put adds a trie pointer to the tries map
func (dth *dataTriesHolder) Put(key []byte, tr temporary.Trie) {
	dth.mutex.Lock()
	dth.tries[string(key)] = tr
	dth.mutex.Unlock()
}

// Replace changes a trie pointer to the tries map
func (dth *dataTriesHolder) Replace(key []byte, tr temporary.Trie) {
	dth.Put(key, tr)
}

// Get returns the trie pointer that is stored in the map at the given key
func (dth *dataTriesHolder) Get(key []byte) temporary.Trie {
	dth.mutex.Lock()
	defer dth.mutex.Unlock()

	return dth.tries[string(key)]
}

// GetAll returns all trie pointers from the map
func (dth *dataTriesHolder) GetAll() []temporary.Trie {
	dth.mutex.Lock()
	defer dth.mutex.Unlock()

	tries := make([]temporary.Trie, 0)
	for _, trie := range dth.tries {
		tries = append(tries, trie)
	}

	return tries
}

// GetAllTries returns the tries with key value map
func (dth *dataTriesHolder) GetAllTries() map[string]temporary.Trie {
	dth.mutex.Lock()
	defer dth.mutex.Unlock()

	copyTries := make(map[string]temporary.Trie, len(dth.tries))
	for key, trie := range dth.tries {
		copyTries[key] = trie
	}

	return copyTries
}

// Reset clears the tries map
func (dth *dataTriesHolder) Reset() {
	dth.mutex.Lock()
	dth.tries = make(map[string]temporary.Trie)
	dth.mutex.Unlock()
}

// IsInterfaceNil returns true if underlying object is nil
func (dth *dataTriesHolder) IsInterfaceNil() bool {
	return dth == nil
}
