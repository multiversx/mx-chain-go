package statistics

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
)

type stateStatistics struct {
	numCache         uint64
	numSyncCache     uint64
	numSnapshotCache uint64

	numPersister         sync.Map
	numSyncPersister     sync.Map
	numSnapshotPersister sync.Map

	numTrie uint64
}

// NewStateStatistics returns a structure able to collect statistics for state
func NewStateStatistics() *stateStatistics {
	return &stateStatistics{}
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

	ss.resetMap(&ss.numPersister)

	atomic.StoreUint64(&ss.numTrie, 0)
}

func (ss *stateStatistics) resetMap(counters *sync.Map) {
	counters.Range(func(k, v interface{}) bool {
		counters.Store(k, uint64(0))
		return true
	})
}

func (ss *stateStatistics) incrMapOp(counters *sync.Map, epoch uint32) {
	val, _ := counters.LoadOrStore(epoch, uint64(0))
	counters.Store(epoch, val.(uint64)+uint64(1))
}

func (ss *stateStatistics) getMapOp(counters *sync.Map, epoch uint32) uint64 {
	val, ok := counters.Load(epoch)
	if !ok {
		return uint64(0)
	}

	return val.(uint64)
}

// Reset will reset the contained values to 0
func (ss *stateStatistics) ResetSync() {
	atomic.StoreUint64(&ss.numSyncCache, ss.CacheOp())

	ss.numPersister.Range(func(k, v interface{}) bool {
		ss.numSyncPersister.Store(k, v)
		return true
	})
}

func (ss *stateStatistics) SyncStats() (uint64, map[uint32]uint64) {
	numSyncCache := ss.CacheOp() - ss.SyncCacheOp()

	persisterStats := make(map[uint32]uint64)

	ss.numPersister.Range(func(k, v interface{}) bool {
		persisterVal, ok := ss.numPersister.Load(k)
		if !ok {
			persisterVal = uint64(0)
		}
		syncPersisterVal, ok := ss.numSyncPersister.Load(k)
		if !ok {
			syncPersisterVal = uint64(0)
		}

		numSyncPersisterPerEpoch := persisterVal.(uint64) - syncPersisterVal.(uint64)
		epoch := k.(uint32)

		persisterStats[epoch] = numSyncPersisterPerEpoch

		return true
	})

	return numSyncCache, persisterStats
}

func (ss *stateStatistics) PrintSync() string {
	numSyncCache, persisterStats := ss.SyncStats()

	stats := make([]string, 0)
	stats = append(stats, fmt.Sprintf("num sync cache op = %v", numSyncCache))

	for epoch, numOp := range persisterStats {
		stats = append(stats, fmt.Sprintf("num persister epoch = %v op = %v", epoch, numOp))
	}

	return strings.Join(stats, " ")
}

// Reset will reset the contained values to 0
func (ss *stateStatistics) ResetSnapshot() {
	atomic.StoreUint64(&ss.numSnapshotCache, 0)

	ss.resetMap(&ss.numSnapshotPersister)
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
	atomic.AddUint64(&ss.numSnapshotCache, 1)
}

// SnapshotCacheOp returns the number of cached operations
func (ss *stateStatistics) SnapshotCacheOp() uint64 {
	return atomic.LoadUint64(&ss.numSnapshotCache)
}

// IncrPersisterOp will increment persister counter
func (ss *stateStatistics) IncrPersisterOp(epoch uint32) {
	ss.incrMapOp(&ss.numPersister, epoch)
}

// PersisterOp returns the number of persister operations
func (ss *stateStatistics) PersisterOp(epoch uint32) uint64 {
	return ss.getMapOp(&ss.numPersister, epoch)
}

// IncrSyncPersisterOp will increment persister counter
func (ss *stateStatistics) IncrSyncPersisterOp(epoch uint32) {
	ss.incrMapOp(&ss.numSyncPersister, epoch)
}

// SyncPersisterOp returns the number of persister operations
func (ss *stateStatistics) SyncPersisterOp(epoch uint32) uint64 {
	return ss.getMapOp(&ss.numSyncPersister, epoch)
}

// IncrSnapshotPersisterOp will increment persister counter
func (ss *stateStatistics) IncrSnapshotPersisterOp(epoch uint32) {
	ss.incrMapOp(&ss.numSnapshotPersister, epoch)
}

// SyncPersisterOp returns the number of persister operations
func (ss *stateStatistics) SnapshotPersisterOp(epoch uint32) uint64 {
	return ss.getMapOp(&ss.numSnapshotPersister, epoch)
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

	ss.numPersister.Range(func(k, v interface{}) bool {
		val, _ := ss.numPersister.Load(k)
		stats = append(stats, fmt.Sprintf("num persister op = %v", val))
		return true
	})

	ss.numSnapshotPersister.Range(func(k, v interface{}) bool {
		val, _ := ss.numSnapshotPersister.Load(k)
		stats = append(stats, fmt.Sprintf("num snapshot persister op = %v", val))
		return true
	})

	stats = append(stats, fmt.Sprintf("max trie op = %v", atomic.LoadUint64(&ss.numTrie)))

	return strings.Join(stats, " ")
}

// IsInterfaceNil returns true if there is no value under the interface
func (ss *stateStatistics) IsInterfaceNil() bool {
	return ss == nil
}
