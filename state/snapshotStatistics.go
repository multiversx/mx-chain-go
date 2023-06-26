package state

import (
	"sync"
	"time"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/trie/statistics"
)

type snapshotStatistics struct {
	trieStatisticsCollector common.TriesStatisticsCollector

	startTime time.Time

	wgSnapshot *sync.WaitGroup
	wgSync     *sync.WaitGroup
	mutex      sync.RWMutex
}

func newSnapshotStatistics(snapshotDelta int, syncDelta int) *snapshotStatistics {
	wgSnapshot := &sync.WaitGroup{}
	wgSnapshot.Add(snapshotDelta)

	wgSync := &sync.WaitGroup{}
	wgSync.Add(syncDelta)
	return &snapshotStatistics{
		wgSnapshot:              wgSnapshot,
		wgSync:                  wgSync,
		startTime:               time.Now(),
		trieStatisticsCollector: statistics.NewTrieStatisticsCollector(),
	}
}

// SnapshotFinished marks the ending of a snapshot goroutine
func (ss *snapshotStatistics) SnapshotFinished() {
	ss.wgSnapshot.Done()
}

// NewSnapshotStarted marks the starting of a new snapshot goroutine
func (ss *snapshotStatistics) NewSnapshotStarted() {
	ss.wgSnapshot.Add(1)
}

// WaitForSnapshotsToFinish will wait until the waitGroup counter is zero
func (ss *snapshotStatistics) WaitForSnapshotsToFinish() {
	ss.wgSnapshot.Wait()
}

// AddTrieStats adds the given trie stats to the snapshot statistics
func (ss *snapshotStatistics) AddTrieStats(trieStats common.TrieStatisticsHandler, trieType common.TrieType) {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	ss.trieStatisticsCollector.Add(trieStats, trieType)
}

// WaitForSyncToFinish will wait until the waitGroup counter is zero
func (ss *snapshotStatistics) WaitForSyncToFinish() {
	ss.wgSync.Wait()
}

// SyncFinished marks the end of the sync process
func (ss *snapshotStatistics) SyncFinished() {
	ss.wgSync.Done()
}

// GetSnapshotDuration will get the duration in seconds of the last snapshot
func (ss *snapshotStatistics) GetSnapshotDuration() int64 {
	duration := time.Since(ss.startTime).Truncate(time.Second)
	return int64(duration.Seconds())
}

// PrintStats will print the stats after the snapshot has finished
func (ss *snapshotStatistics) PrintStats(identifier string, rootHash []byte) {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()

	log.Debug("snapshot statistics",
		"type", identifier,
		"duration", time.Since(ss.startTime).Truncate(time.Second),
		"rootHash", rootHash,
	)
	ss.trieStatisticsCollector.Print()
}

// GetSnapshotNumNodes returns the number of nodes from the snapshot
func (ss *snapshotStatistics) GetSnapshotNumNodes() uint64 {
	return ss.trieStatisticsCollector.GetNumNodes()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ss *snapshotStatistics) IsInterfaceNil() bool {
	return ss == nil
}
