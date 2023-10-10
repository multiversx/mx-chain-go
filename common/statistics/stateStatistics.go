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

// ResetAll will reset all statistics
func (ss *stateStatistics) ResetAll() {
	ss.Reset()
	ss.ResetSync()
	ss.ResetSnapshot()
}

// Reset will reset processing statistics
func (ss *stateStatistics) Reset() {
	atomic.StoreUint64(&ss.numCache, 0)

	ss.resetMap(&ss.numPersister)

	atomic.StoreUint64(&ss.numTrie, 0)
}

// ResetSync will set sync statistics based on processing statistics
func (ss *stateStatistics) ResetSync() {
	atomic.StoreUint64(&ss.numSyncCache, ss.Cache())

	ss.numPersister.Range(func(k, v interface{}) bool {
		ss.numSyncPersister.Store(k, v)
		return true
	})
}

func (ss *stateStatistics) getSyncStats() (uint64, map[uint32]uint64) {
	numSyncCache := ss.Cache() - ss.SyncCache()

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

// ResetSnapshot will reset snapshot statistics
func (ss *stateStatistics) ResetSnapshot() {
	atomic.StoreUint64(&ss.numSnapshotCache, 0)
	ss.resetMap(&ss.numSnapshotPersister)
}

// IncrCache will increment cache counter
func (ss *stateStatistics) IncrCache() {
	atomic.AddUint64(&ss.numCache, 1)
}

// Cache returns the number of cached operations
func (ss *stateStatistics) Cache() uint64 {
	return atomic.LoadUint64(&ss.numCache)
}

// IncrSyncCache will increment sync cache counter
func (ss *stateStatistics) IncrSyncCache() {
	atomic.AddUint64(&ss.numSyncCache, 1)
}

// SyncCache returns the number of sync cached operations
func (ss *stateStatistics) SyncCache() uint64 {
	return atomic.LoadUint64(&ss.numSyncCache)
}

// IncrSnapshotCache will increment snapshot cache counter
func (ss *stateStatistics) IncrSnapshotCache() {
	atomic.AddUint64(&ss.numSnapshotCache, 1)
}

// SnapshotCache returns the number of snapshot  cached operations
func (ss *stateStatistics) SnapshotCache() uint64 {
	return atomic.LoadUint64(&ss.numSnapshotCache)
}

// IncrPersister will increment persister counter
func (ss *stateStatistics) IncrPersister(epoch uint32) {
	ss.incrMapOp(&ss.numPersister, epoch)
}

// Persister returns the number of persister operations
func (ss *stateStatistics) Persister(epoch uint32) uint64 {
	return ss.getMapOp(&ss.numPersister, epoch)
}

// IncrSyncPersister will increment sync persister counter
func (ss *stateStatistics) IncrSyncPersister(epoch uint32) {
	ss.incrMapOp(&ss.numSyncPersister, epoch)
}

// SyncPersister returns the number of sync persister operations
func (ss *stateStatistics) SyncPersister(epoch uint32) uint64 {
	return ss.getMapOp(&ss.numSyncPersister, epoch)
}

// IncrSnapshotPersister will increment snapshot persister counter
func (ss *stateStatistics) IncrSnapshotPersister(epoch uint32) {
	ss.incrMapOp(&ss.numSnapshotPersister, epoch)
}

// SnapshotPersister returns the number of snapshot persister operations
func (ss *stateStatistics) SnapshotPersister(epoch uint32) uint64 {
	return ss.getMapOp(&ss.numSnapshotPersister, epoch)
}

// IncrTrie will increment trie counter
func (ss *stateStatistics) IncrTrie() {
	atomic.AddUint64(&ss.numTrie, 1)
}

// Trie returns the number of trie operations
func (ss *stateStatistics) Trie() uint64 {
	return atomic.LoadUint64(&ss.numTrie)
}

// SnapshotStats returns collected snapshot statistics as string
func (ss *stateStatistics) SnapshotStats() string {
	stats := make([]string, 0)

	stats = append(stats, fmt.Sprintf("num snapshot cache op = %v", atomic.LoadUint64(&ss.numSnapshotCache)))

	ss.numSnapshotPersister.Range(func(k, v interface{}) bool {
		val, _ := ss.numSnapshotPersister.Load(k)
		stats = append(stats, fmt.Sprintf("num snapshot persister op = %v", val))
		return true
	})

	return strings.Join(stats, " ")
}

// SnapshotStats returns collected processing statistics as string
func (ss *stateStatistics) ProcessingStats() string {
	stats := make([]string, 0)

	stats = append(stats, fmt.Sprintf("num cache op = %v", atomic.LoadUint64(&ss.numCache)))

	ss.numPersister.Range(func(k, v interface{}) bool {
		val, _ := ss.numPersister.Load(k)
		stats = append(stats, fmt.Sprintf("num persister op = %v", val))
		return true
	})

	stats = append(stats, fmt.Sprintf("max trie op = %v", atomic.LoadUint64(&ss.numTrie)))

	return strings.Join(stats, " ")
}

// SnapshotStats returns sync processing statistics as string
func (ss *stateStatistics) SyncStats() string {
	numSyncCache, persisterStats := ss.getSyncStats()

	stats := make([]string, 0)
	stats = append(stats, fmt.Sprintf("num sync cache op = %v", numSyncCache))

	for epoch, numOp := range persisterStats {
		stats = append(stats, fmt.Sprintf("num persister epoch = %v op = %v", epoch, numOp))
	}

	return strings.Join(stats, " ")
}

// IsInterfaceNil returns true if there is no value under the interface
func (ss *stateStatistics) IsInterfaceNil() bool {
	return ss == nil
}
