package state

import (
	"sync"
	"testing"

	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/trie/statistics"
)

func TestSnapshotStatistics_Concurrency(t *testing.T) {
	t.Parallel()

	wg := &sync.WaitGroup{}
	stats := statistics.NewTrieStatisticsCollector()
	ss := &snapshotStatistics{
		wgSnapshot:              wg,
		trieStatisticsCollector: stats,
	}

	numRuns := 100
	for i := 0; i < numRuns; i++ {
		ss.NewSnapshotStarted()
		go func() {
			ss.AddTrieStats(getTrieStatsDTO(5, 60).GetTrieStats())
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
