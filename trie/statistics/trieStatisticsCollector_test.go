package statistics

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/stretchr/testify/assert"
)

func TestSnapshotStatistics_Add(t *testing.T) {
	t.Parallel()

	tsc := NewTrieStatisticsCollector()

	tsc.Add(nil, common.MainTrie) // coverage, early exit

	numInserts := 100
	for i := 0; i < numInserts; i++ {
		tsc.Add(getTrieStats(rand.Intn(numInserts), uint64(rand.Intn(numInserts))), common.DataTrie)
		isSortedBySize := sort.SliceIsSorted(tsc.triesBySize, func(a, b int) bool {
			if tsc.triesBySize[b] == nil && tsc.triesBySize[a] == nil {
				return false
			}
			if tsc.triesBySize[a] == nil {
				return false
			}

			return tsc.triesBySize[b].GetTotalNodesSize() < tsc.triesBySize[a].GetTotalNodesSize()
		})

		isSortedByDepth := sort.SliceIsSorted(tsc.triesByDepth, func(a, b int) bool {
			if tsc.triesByDepth[b] == nil && tsc.triesByDepth[a] == nil {
				return false
			}
			if tsc.triesByDepth[a] == nil {
				return false
			}

			return tsc.triesByDepth[b].GetMaxTrieDepth() < tsc.triesByDepth[a].GetMaxTrieDepth()
		})

		assert.True(t, isSortedBySize)
		assert.True(t, isSortedByDepth)
		assert.Equal(t, numTriesToPrint, len(tsc.triesBySize))
		assert.Equal(t, numTriesToPrint, len(tsc.triesByDepth))
		assert.Equal(t, uint64(i+1), tsc.GetNumNodes())

		tsc.Print() // coverage
	}
}

func getTrieStats(maxLevel int, size uint64) common.TrieStatisticsHandler {
	ts := NewTrieStatistics()
	ts.AddBranchNode(maxLevel, size)

	return ts
}

func TestGetNumTriesByTypeString(t *testing.T) {
	t.Parallel()

	numMainTries := 1
	numDataTries := 500
	numTriesByType := make(map[common.TrieType]uint64)
	numTriesByType[common.MainTrie] = uint64(numMainTries)
	numTriesByType[common.DataTrie] = uint64(numDataTries)

	numTriesByTypeString := getNumTriesByTypeString(numTriesByType)
	expectedRes := fmt.Sprintf("%v: %v, %v: %v", common.DataTrie, numDataTries, common.MainTrie, numMainTries)
	assert.Equal(t, expectedRes, numTriesByTypeString)
}
