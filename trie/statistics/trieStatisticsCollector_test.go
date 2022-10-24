package statistics

import (
	"math/rand"
	"sort"
	"testing"

	"github.com/ElrondNetwork/elrond-go/common/disabled"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSnapshotStatistics_AddTrieStats(t *testing.T) {
	t.Parallel()

	tsc, err := NewTrieStatisticsCollector(disabled.NewAppStatusHandler())
	require.Nil(t, err)

	numInserts := 100
	for i := 0; i < numInserts; i++ {
		tsc.Add(getTrieStatsDTO(rand.Intn(numInserts), uint64(rand.Intn(numInserts))))
		isSortedBySize := sort.SliceIsSorted(tsc.triesBySize, func(a, b int) bool {
			if tsc.triesBySize[b] == nil && tsc.triesBySize[a] == nil {
				return false
			}
			if tsc.triesBySize[a] == nil {
				return false
			}

			return tsc.triesBySize[b].TotalNodesSize < tsc.triesBySize[a].TotalNodesSize
		})

		isSortedByDepth := sort.SliceIsSorted(tsc.triesByDepth, func(a, b int) bool {
			if tsc.triesByDepth[b] == nil && tsc.triesByDepth[a] == nil {
				return false
			}
			if tsc.triesByDepth[a] == nil {
				return false
			}

			return tsc.triesByDepth[b].MaxTrieDepth < tsc.triesByDepth[a].MaxTrieDepth
		})

		assert.True(t, isSortedBySize)
		assert.True(t, isSortedByDepth)
		assert.Equal(t, numTriesToPrint, len(tsc.triesBySize))
		assert.Equal(t, numTriesToPrint, len(tsc.triesByDepth))
	}
}

func getTrieStatsDTO(maxLevel int, size uint64) *TrieStatsDTO {
	ts := NewTrieStatistics()
	ts.AddBranchNode(maxLevel, size)
	return ts.GetTrieStats()
}
