package statistics

import (
	"fmt"
	"strings"
	"sync/atomic"
)

type stateStatistics struct {
	numCache     uint64
	numPersister uint64
	numTrie      uint64
}

// NewStateStatistics returns a structure able to collect statistics for state
func NewStateStatistics() *stateStatistics {
	return &stateStatistics{}
}

// Reset will reset the contained values to 0
func (ss *stateStatistics) Reset() {
	atomic.StoreUint64(&ss.numCache, 0)
	atomic.StoreUint64(&ss.numPersister, 0)
	atomic.StoreUint64(&ss.numTrie, 0)
}

// IncrCacheOp will increment cache counter
func (ss *stateStatistics) IncrCacheOp() {
	atomic.AddUint64(&ss.numCache, 1)
}

// CacheOp returns the number of cached operations
func (ss *stateStatistics) CacheOp() uint64 {
	return atomic.LoadUint64(&ss.numCache)
}

// IncrPersisterOp will increment persister counter
func (ss *stateStatistics) IncrPersisterOp() {
	atomic.AddUint64(&ss.numPersister, 1)
}

// PersisterOp returns the number of persister operations
func (ss *stateStatistics) PersisterOp() uint64 {
	return atomic.LoadUint64(&ss.numPersister)
}

// IncrTrieOp will increment trie counter
func (ss *stateStatistics) IncrTrieOp() {
	atomic.AddUint64(&ss.numTrie, 1)
}

// TrieOp returns the number of trie operations
func (ss *stateStatistics) TrieOp() uint64 {
	return atomic.LoadUint64(&ss.numTrie)
}

// ToString returns collected statistics as string
func (ss *stateStatistics) ToString() string {
	stats := make([]string, 0)

	stats = append(stats, fmt.Sprintf("num cache op = %v", atomic.LoadUint64(&ss.numCache)))
	stats = append(stats, fmt.Sprintf("num persister op = %v", atomic.LoadUint64(&ss.numPersister)))
	stats = append(stats, fmt.Sprintf("max trie op = %v", atomic.LoadUint64(&ss.numTrie)))

	return strings.Join(stats, " ")
}

// IsInterfaceNil returns true if there is no value under the interface
func (ss *stateStatistics) IsInterfaceNil() bool {
	return ss == nil
}
