package statistics

import (
	"strconv"
	"strings"

	"github.com/ElrondNetwork/elrond-go-core/core"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("trieStatistics")

const numTriesToPrint = 10

type trieStatisticsCollector struct {
	numNodes     uint64
	numDataTries uint64
	triesSize    uint64
	triesBySize  []*TrieStatsDTO
	triesByDepth []*TrieStatsDTO
}

// NewTrieStatisticsCollector creates a new instance of trieStatisticsCollector
func NewTrieStatisticsCollector() *trieStatisticsCollector {
	return &trieStatisticsCollector{
		numNodes:     0,
		numDataTries: 0,
		triesSize:    0,
		triesBySize:  make([]*TrieStatsDTO, numTriesToPrint),
		triesByDepth: make([]*TrieStatsDTO, numTriesToPrint),
	}
}

// Add adds the given trie statistics to the statistics collector
func (tsc *trieStatisticsCollector) Add(trieStats *TrieStatsDTO) {
	tsc.numNodes += trieStats.TotalNumNodes
	tsc.triesSize += trieStats.TotalNodesSize
	tsc.numDataTries++

	insertInSortedArray(tsc.triesBySize, trieStats, isLessSize)
	insertInSortedArray(tsc.triesByDepth, trieStats, isLessDeep)
}

// Print will print all the collected statistics
func (tsc *trieStatisticsCollector) Print() {
	triesBySize := " \n top " + strconv.Itoa(numTriesToPrint) + " tries by size \n"
	triesByDepth := " \n top " + strconv.Itoa(numTriesToPrint) + " tries by depth \n"

	log.Debug("tries statistics",
		"num of nodes", tsc.numNodes,
		"total size", core.ConvertBytes(tsc.triesSize),
		"num tries", tsc.numDataTries,
		triesBySize, getOrderedTries(tsc.triesBySize),
		triesByDepth, getOrderedTries(tsc.triesByDepth),
	)
}

// GetNumNodes returns the number of nodes
func (tsc *trieStatisticsCollector) GetNumNodes() uint64 {
	return tsc.numNodes
}

func getOrderedTries(tries []*TrieStatsDTO) string {
	triesStats := make([]string, 0)
	for i := 0; i < len(tries); i++ {
		if tries[i] == nil {
			continue
		}
		triesStats = append(triesStats, strings.Join(tries[i].ToString(), " "))
	}

	return strings.Join(triesStats, "\n")
}

func isLessSize(a *TrieStatsDTO, b *TrieStatsDTO) bool {
	return a.TotalNodesSize < b.TotalNodesSize
}

func isLessDeep(a *TrieStatsDTO, b *TrieStatsDTO) bool {
	return a.MaxTrieDepth < b.MaxTrieDepth
}

func insertInSortedArray(
	array []*TrieStatsDTO,
	ts *TrieStatsDTO,
	isLess func(*TrieStatsDTO, *TrieStatsDTO) bool,
) {
	insertIndex := numTriesToPrint
	lastNilIndex := numTriesToPrint
	for i := numTriesToPrint - 1; i >= 0; i-- {
		currentTrie := array[i]
		if currentTrie == nil {
			lastNilIndex = i
			if i == 0 {
				array[i] = ts
			}
			continue
		}

		if isLess(currentTrie, ts) {
			insertIndex = i
			continue
		}

		break
	}

	if insertIndex < numTriesToPrint {
		array = append(array[:insertIndex+1], array[insertIndex:numTriesToPrint-1]...)
		array[insertIndex] = ts
		return
	}

	if lastNilIndex < numTriesToPrint {
		array[lastNilIndex] = ts
	}
}
