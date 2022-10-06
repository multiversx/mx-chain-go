package state

import (
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
)

type snapshotStatistics struct {
	numNodes     uint64
	numDataTries uint64
	trieSize     uint64
	startTime    time.Time

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
		wgSnapshot: wgSnapshot,
		wgSync:     wgSync,
		startTime:  time.Now(),
	}
}

// AddSize will add the given size to the trie size counter
func (ss *snapshotStatistics) AddSize(size uint64) {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	ss.numNodes++
	ss.trieSize += size
}

// SnapshotFinished marks the ending of a snapshot goroutine
func (ss *snapshotStatistics) SnapshotFinished() {
	ss.wgSnapshot.Done()
}

// NewSnapshotStarted marks the starting of a new snapshot goroutine
func (ss *snapshotStatistics) NewSnapshotStarted() {
	ss.wgSnapshot.Add(1)
}

// NewDataTrie increases the data Tries counter
func (ss *snapshotStatistics) NewDataTrie() {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	ss.numDataTries++
}

// WaitForSnapshotsToFinish will wait until the waitGroup counter is zero
func (ss *snapshotStatistics) WaitForSnapshotsToFinish() {
	ss.wgSnapshot.Wait()
}

// WaitForSyncToFinish will wait until the waitGroup counter is zero
func (ss *snapshotStatistics) WaitForSyncToFinish() {
	ss.wgSync.Wait()
}

// SyncFinished marks the end of the sync process
func (ss *snapshotStatistics) SyncFinished() {
	ss.wgSync.Done()
}

// GetUnixStartTime will get the start time in unix format
func (ss *snapshotStatistics) GetUnixStartTime() int64 {
	return ss.startTime.Unix()
}

// PrintStats will print the stats after the snapshot has finished
func (ss *snapshotStatistics) PrintStats(identifier string, rootHash []byte) {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()

	log.Debug("snapshot statistics",
		"type", identifier,
		"duration", time.Since(ss.startTime).Truncate(time.Second),
		"num of nodes copied", ss.numNodes,
		"total size copied", core.ConvertBytes(ss.trieSize),
		"num data tries copied", ss.numDataTries,
		"rootHash", rootHash,
	)
}
