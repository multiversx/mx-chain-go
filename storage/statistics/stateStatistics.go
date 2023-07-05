package statistics

import (
	"sync"
)

type stateStatistics struct {
	sync.RWMutex
	numCache     int
	numPersister int
	numTrie      int
}

// NewStateStatistics returns a structure able to collect statistics for state
func NewStateStatistics() *stateStatistics {
	return &stateStatistics{}
}

// Reset will reset the contained values to 0
func (tss *stateStatistics) Reset() {
	tss.Lock()
	tss.Unlock()
}

// AddNumCache will add the provided value to the cache counter
func (ss *stateStatistics) AddNumCache(value int) {
	ss.Lock()
	ss.numCache += value
	ss.Unlock()
}

// NumCache returns the number of cached keys
func (ss *stateStatistics) NumCache() int {
	ss.RLock()
	defer ss.RUnlock()

	return ss.numCache
}

// AddNumPersister will add the provided value to the persister counter
func (ss *stateStatistics) AddNumPersister(value int) {
	ss.Lock()
	ss.numPersister += value
	ss.Unlock()
}

// NumPersister returns the number of persister keys
func (ss *stateStatistics) NumPersister() int {
	ss.RLock()
	defer ss.RUnlock()

	return ss.numPersister
}

// AddNumTrie will add the provided value to the trie counter
func (ss *stateStatistics) AddNumTrie(value int) {
	ss.Lock()
	ss.numTrie += value
	ss.Unlock()
}

// NumTrie returns the number of trie keys
func (ss *stateStatistics) NumTrie() int {
	ss.RLock()
	defer ss.RUnlock()

	return ss.numTrie
}

// IsInterfaceNil returns true if there is no value under the interface
func (ss *stateStatistics) IsInterfaceNil() bool {
	return ss == nil
}
