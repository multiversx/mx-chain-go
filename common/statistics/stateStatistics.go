package statistics

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type stateStatistics struct {
	numCache         uint64
	numSnapshotCache uint64

	numPersister         map[uint32]uint64
	numSnapshotPersister map[uint32]uint64
	mutPersisters        sync.RWMutex

	numTrie uint64
}

// NewStateStatistics returns a structure able to collect statistics for state
func NewStateStatistics() *stateStatistics {
	return &stateStatistics{
		numPersister:         make(map[uint32]uint64),
		numSnapshotPersister: make(map[uint32]uint64),
	}
}

// ResetAll will reset all statistics
func (ss *stateStatistics) ResetAll() {
	ss.Reset()
	ss.ResetSnapshot()
}

// Reset will reset processing statistics
func (ss *stateStatistics) Reset() {
	atomic.StoreUint64(&ss.numCache, 0)

	ss.mutPersisters.Lock()
	ss.numPersister = make(map[uint32]uint64)
	ss.mutPersisters.Unlock()

	atomic.StoreUint64(&ss.numTrie, 0)
}

// ResetSnapshot will reset snapshot statistics
func (ss *stateStatistics) ResetSnapshot() {
	atomic.StoreUint64(&ss.numSnapshotCache, 0)

	ss.mutPersisters.Lock()
	ss.numSnapshotPersister = make(map[uint32]uint64)
	ss.mutPersisters.Unlock()
}

// IncrementCache will increment cache counter
func (ss *stateStatistics) IncrementCache() {
	atomic.AddUint64(&ss.numCache, 1)
}

// Cache returns the number of cached operations
func (ss *stateStatistics) Cache() uint64 {
	return atomic.LoadUint64(&ss.numCache)
}

// IncrementSnapshotCache will increment snapshot cache counter
func (ss *stateStatistics) IncrementSnapshotCache() {
	atomic.AddUint64(&ss.numSnapshotCache, 1)
}

// SnapshotCache returns the number of snapshot  cached operations
func (ss *stateStatistics) SnapshotCache() uint64 {
	return atomic.LoadUint64(&ss.numSnapshotCache)
}

// IncrementPersister will increment persister counter
func (ss *stateStatistics) IncrementPersister(epoch uint32) {
	ss.mutPersisters.Lock()
	defer ss.mutPersisters.Unlock()

	ss.numPersister[epoch]++
}

// Persister returns the number of persister operations
func (ss *stateStatistics) Persister(epoch uint32) uint64 {
	ss.mutPersisters.RLock()
	defer ss.mutPersisters.RUnlock()

	return ss.numPersister[epoch]
}

// IncrementSnapshotPersister will increment snapshot persister counter
func (ss *stateStatistics) IncrementSnapshotPersister(epoch uint32) {
	ss.mutPersisters.Lock()
	defer ss.mutPersisters.Unlock()

	ss.numSnapshotPersister[epoch]++
}

// SnapshotPersister returns the number of snapshot persister operations
func (ss *stateStatistics) SnapshotPersister(epoch uint32) uint64 {
	ss.mutPersisters.RLock()
	defer ss.mutPersisters.RUnlock()

	return ss.numSnapshotPersister[epoch]
}

// IncrementTrie will increment trie counter
func (ss *stateStatistics) IncrementTrie() {
	atomic.AddUint64(&ss.numTrie, 1)
}

// Trie returns the number of trie operations
func (ss *stateStatistics) Trie() uint64 {
	return atomic.LoadUint64(&ss.numTrie)
}

// SnapshotStats returns collected snapshot statistics as string
func (ss *stateStatistics) SnapshotStats() []string {
	stats := make([]string, 0)

	stats = append(stats, fmt.Sprintf("snapshot cache op = %v", atomic.LoadUint64(&ss.numSnapshotCache)))

	ss.mutPersisters.RLock()
	defer ss.mutPersisters.RUnlock()

	for epoch, counter := range ss.numSnapshotPersister {
		stats = append(stats, fmt.Sprintf("snapshot persister epoch = %v op = %v", epoch, counter))
	}

	return stats
}

// ProcessingStats returns collected processing statistics as string
func (ss *stateStatistics) ProcessingStats() []string {
	stats := make([]string, 0)

	stats = append(stats, fmt.Sprintf("cache op = %v", atomic.LoadUint64(&ss.numCache)))

	ss.mutPersisters.RLock()
	defer ss.mutPersisters.RUnlock()

	for epoch, counter := range ss.numPersister {
		stats = append(stats, fmt.Sprintf("persister epoch = %v op = %v", epoch, counter))
	}

	stats = append(stats, fmt.Sprintf("trie op = %v", atomic.LoadUint64(&ss.numTrie)))

	return stats
}

// IsInterfaceNil returns true if there is no value under the interface
func (ss *stateStatistics) IsInterfaceNil() bool {
	return ss == nil
}
