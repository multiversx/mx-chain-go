package state

import (
	"sync"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/trie/statistics"
)

func TestSnapshotStatistics_Concurrency(t *testing.T) {
	t.Parallel()

	wg := &sync.WaitGroup{}
	ss := &snapshotStatistics{
		wgSnapshot:              wg,
		trieStatisticsCollector: statistics.NewTrieStatisticsCollector(),
	}

	numRuns := 100
	for i := 0; i < numRuns; i++ {
		ss.NewSnapshotStarted()
		go func() {
			ss.AddTrieStats(getTrieStatsDTO(5, 60), common.DataTrie)
			ss.SnapshotFinished()
		}()
	}

	wg.Wait()
}

func getTrieStatsDTO(maxLevel int, size uint64) common.TrieStatisticsHandler {
	ts := statistics.NewTrieStatistics()
	ts.AddBranchNode(maxLevel, size)
	return ts
}
