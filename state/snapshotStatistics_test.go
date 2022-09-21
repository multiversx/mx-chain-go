package state

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSnapshotStatistics_AddSize(t *testing.T) {
	ss := &snapshotStatistics{}
	assert.Equal(t, uint64(0), ss.numNodes)
	assert.Equal(t, uint64(0), ss.trieSize)

	ss.AddSize(8)
	ss.AddSize(16)
	ss.AddSize(32)
	assert.Equal(t, uint64(3), ss.numNodes)
	assert.Equal(t, uint64(56), ss.trieSize)
}

func TestSnapshotStatistics_Concurrency(t *testing.T) {
	wg := &sync.WaitGroup{}
	ss := &snapshotStatistics{
		wgSnapshot: wg,
	}

	numRuns := 100
	for i := 0; i < numRuns; i++ {
		ss.NewSnapshotStarted()
		go func() {
			ss.AddSize(10)
			ss.NewDataTrie()
			ss.SnapshotFinished()
		}()
	}

	wg.Wait()
	assert.Equal(t, uint64(100), ss.numNodes)
	assert.Equal(t, uint64(1000), ss.trieSize)
	assert.Equal(t, uint64(100), ss.numDataTries)
}
