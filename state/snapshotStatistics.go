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

	wg    *sync.WaitGroup
	mutex sync.RWMutex
}

func newSnapshotStatistics(delta int) *snapshotStatistics {
	wg := &sync.WaitGroup{}
	wg.Add(delta)
	return &snapshotStatistics{
		wg:        wg,
		startTime: time.Now(),
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
	ss.wg.Done()
}

// NewSnapshotStarted marks the starting of a new snapshot goroutine
func (ss *snapshotStatistics) NewSnapshotStarted() {
	ss.wg.Add(1)
}

// NewDataTrie increases the data Tries counter
func (ss *snapshotStatistics) NewDataTrie() {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	ss.numDataTries++
}

// WaitForSnapshotsToFinish will wait until the waitGroup counter is zero
func (ss *snapshotStatistics) WaitForSnapshotsToFinish() {
	ss.wg.Wait()
}
