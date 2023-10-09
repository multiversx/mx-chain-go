package statistics

import (
	"fmt"
	"strings"
	"sync/atomic"
)

type stateStatistics struct {
	// TODO: add multiple stats types, maybe inner components
	// TODO: use mutex instead of sync/atomic
	numCache         uint64
	numSyncCache     uint64
	numSnapshotCache uint64

	numPersister         map[uint32]uint64
	numSyncPersister     map[uint32]uint64
	numSnapshotPersister map[uint32]uint64

	numTrie uint64
}

// NewStateStatistics returns a structure able to collect statistics for state
func NewStateStatistics() *stateStatistics {
	return &stateStatistics{
		numPersister:         make(map[uint32]uint64),
		numSyncPersister:     make(map[uint32]uint64),
		numSnapshotPersister: make(map[uint32]uint64),
	}
}

// ResetAll will reset all statistics
func (ss *stateStatistics) ResetAll() {
	ss.Reset()
	ss.ResetSync()
	ss.ResetSnapshot()
}

// Reset will reset the contained values to 0
func (ss *stateStatistics) Reset() {
	atomic.StoreUint64(&ss.numCache, 0)

	for epoch := range ss.numPersister {
		ss.numPersister[epoch] = 0
	}

	atomic.StoreUint64(&ss.numTrie, 0)
}

// Reset will reset the contained values to 0
func (ss *stateStatistics) ResetSync() {
	atomic.StoreUint64(&ss.numSyncCache, ss.CacheOp())

	for epoch := range ss.numSyncPersister {
		ss.numSyncPersister[epoch] = ss.numPersister[epoch]
	}
}

func (ss *stateStatistics) PrintSync() string {
	numSyncCache := ss.CacheOp() - ss.SyncCacheOp()

	stats := make([]string, 0)
	stats = append(stats, fmt.Sprintf("num sync cache op = %v", numSyncCache))

	for epoch := range ss.numSyncPersister {
		numSyncPersisterPerEpoch := ss.numPersister[epoch] - ss.numSyncPersister[epoch]
		stats = append(stats, fmt.Sprintf("num persister epoch = %v op = %v", epoch, numSyncPersisterPerEpoch))
	}

	return strings.Join(stats, " ")
}

// Reset will reset the contained values to 0
func (ss *stateStatistics) ResetSnapshot() {
	atomic.StoreUint64(&ss.numSnapshotCache, 0)

	for epoch := range ss.numSnapshotPersister {
		ss.numSnapshotPersister[epoch] = 0
	}
}

// IncrCacheOp will increment cache counter
func (ss *stateStatistics) IncrCacheOp() {
	atomic.AddUint64(&ss.numCache, 1)
}

// CacheOp returns the number of cached operations
func (ss *stateStatistics) CacheOp() uint64 {
	return atomic.LoadUint64(&ss.numCache)
}

// IncrSyncCacheOp will increment cache counter
func (ss *stateStatistics) IncrSyncCacheOp() {
	atomic.AddUint64(&ss.numSyncCache, 1)
}

// SyncCacheOp returns the number of cached operations
func (ss *stateStatistics) SyncCacheOp() uint64 {
	return atomic.LoadUint64(&ss.numSyncCache)
}

// IncrSnapshotCacheOp will increment cache counter
func (ss *stateStatistics) IncrSnapshotCacheOp() {
	atomic.AddUint64(&ss.numCache, 1)
}

// SnapshotCacheOp returns the number of cached operations
func (ss *stateStatistics) SnapshotCacheOp() uint64 {
	return atomic.LoadUint64(&ss.numCache)
}

// IncrPersisterOp will increment persister counter
func (ss *stateStatistics) IncrPersisterOp(epoch uint32) {
	ss.numPersister[epoch]++
}

// PersisterOp returns the number of persister operations
func (ss *stateStatistics) PersisterOp(epoch uint32) uint64 {
	return ss.numPersister[epoch]
}

// IncrSyncPersisterOp will increment persister counter
func (ss *stateStatistics) IncrSyncPersisterOp(epoch uint32) {
	ss.numSyncPersister[epoch]++
}

// SyncPersisterOp returns the number of persister operations
func (ss *stateStatistics) SyncPersisterOp(epoch uint32) uint64 {
	return ss.numSyncPersister[epoch]
}

// IncrSnapshotPersisterOp will increment persister counter
func (ss *stateStatistics) IncrSnapshotPersisterOp(epoch uint32) {
	ss.numSnapshotPersister[epoch]++
}

// SyncPersisterOp returns the number of persister operations
func (ss *stateStatistics) SnapshotPersisterOp(epoch uint32) uint64 {
	return ss.numSnapshotPersister[epoch]
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
	stats = append(stats, fmt.Sprintf("num snapshot cache op = %v", atomic.LoadUint64(&ss.numSnapshotCache)))

	for epoch := range ss.numPersister {
		stats = append(stats, fmt.Sprintf("num persister op = %v", ss.numPersister[epoch]))
	}
	for epoch := range ss.numSnapshotPersister {
		stats = append(stats, fmt.Sprintf("num snapshot persister op = %v", ss.numSnapshotPersister[epoch]))
	}

	stats = append(stats, fmt.Sprintf("max trie op = %v", atomic.LoadUint64(&ss.numTrie)))

	return strings.Join(stats, " ")
}

// IsInterfaceNil returns true if there is no value under the interface
func (ss *stateStatistics) IsInterfaceNil() bool {
	return ss == nil
}
