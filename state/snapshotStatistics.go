package state

import (
	"sync"
	"time"
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
