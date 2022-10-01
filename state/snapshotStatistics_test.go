package state

import (
	"math/rand"
	"sort"
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/trie/statistics"
	"github.com/stretchr/testify/assert"
)

func TestSnapshotStatistics_AddSize(t *testing.T) {
	t.Parallel()

	ss := &snapshotStatistics{}
	assert.Equal(t, uint64(0), ss.numNodes)
	assert.Equal(t, uint64(0), ss.triesSize)

	ss.AddSize(8)
	ss.AddSize(16)
	ss.AddSize(32)
	assert.Equal(t, uint64(3), ss.numNodes)
	assert.Equal(t, uint64(56), ss.triesSize)
}

func TestSnapshotStatistics_Concurrency(t *testing.T) {
	t.Parallel()

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
	assert.Equal(t, uint64(1000), ss.triesSize)
	assert.Equal(t, uint64(100), ss.numDataTries)
}

func TestSnapshotStatistics_AddTrieStats(t *testing.T) {
	t.Parallel()

	ss := newSnapshotStatistics(0, 0)

	numInserts := 100
	for i := 0; i < numInserts; i++ {
		ss.AddTrieStats(getTrieStatsDTO(rand.Intn(numInserts), uint64(rand.Intn(numInserts))))
		isSortedBySize := sort.SliceIsSorted(ss.triesBySize, func(a, b int) bool {
			if ss.triesBySize[b] == nil && ss.triesBySize[a] == nil {
				return false
			}
			if ss.triesBySize[a] == nil {
				return false
			}

			return ss.triesBySize[b].TotalNodesSize < ss.triesBySize[a].TotalNodesSize
		})

		isSortedByDepth := sort.SliceIsSorted(ss.triesByDepth, func(a, b int) bool {
			if ss.triesByDepth[b] == nil && ss.triesByDepth[a] == nil {
				return false
			}
			if ss.triesByDepth[a] == nil {
				return false
			}

			return ss.triesByDepth[b].MaxTrieDepth < ss.triesByDepth[a].MaxTrieDepth
		})

		assert.True(t, isSortedBySize)
		assert.True(t, isSortedByDepth)
		assert.Equal(t, numTriesToPrint, len(ss.triesBySize))
		assert.Equal(t, numTriesToPrint, len(ss.triesByDepth))
	}
}

func getTrieStatsDTO(maxLevel int, size uint64) common.TrieStatisticsHandler {
	ts := statistics.NewTrieStatistics()
	ts.AddBranchNode(maxLevel, size)
	return ts
}
