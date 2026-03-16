package triesHolder

import (
	"sync"

	"github.com/multiversx/mx-chain-go/common"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type triesHolder struct {
	tries map[string]common.Trie
	mutex sync.RWMutex
}

// NewTriesHolder creates a new instance of triesHolder
func NewTriesHolder() *triesHolder {
	return &triesHolder{
		tries: make(map[string]common.Trie),
	}
}

// Put adds a trie pointer to the tries map
func (dth *triesHolder) Put(key []byte, tr common.Trie) {
	log.Trace("put trie in tries holder", "key", key)

	dth.mutex.Lock()
	dth.tries[string(key)] = tr
	dth.mutex.Unlock()
}

// Get returns the trie pointer that is stored in the map at the given key
func (dth *triesHolder) Get(key []byte) common.Trie {
	dth.mutex.Lock()
	defer dth.mutex.Unlock()

	return dth.tries[string(key)]
}

// GetAll returns all trie pointers from the map
func (dth *triesHolder) GetAll() []common.Trie {
	dth.mutex.Lock()
	defer dth.mutex.Unlock()

	tries := make([]common.Trie, 0)
	for _, trie := range dth.tries {
		tries = append(tries, trie)
	}

	return tries
}

// Reset clears the tries map
func (dth *triesHolder) Reset() {
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
func (dth *triesHolder) IsInterfaceNil() bool {
	return dth == nil
}
