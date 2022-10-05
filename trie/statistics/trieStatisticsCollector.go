package statistics

import (
	"strconv"
	"strings"

	"github.com/ElrondNetwork/elrond-go-core/core"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("trieStatistics")

type trieStatisticsCollector struct {
	numNodes     uint64
	numDataTries uint64
	triesSize    uint64
	triesBySize  []*TrieStatsDTO
	triesByDepth []*TrieStatsDTO

	numTriesToPrint int
}

func NewTrieStatisticsCollector(numTriesToPrint int) *trieStatisticsCollector {
	return &trieStatisticsCollector{
		numNodes:        0,
		numDataTries:    0,
		triesSize:       0,
		triesBySize:     make([]*TrieStatsDTO, numTriesToPrint),
		triesByDepth:    make([]*TrieStatsDTO, numTriesToPrint),
		numTriesToPrint: numTriesToPrint,
	}
}

func (tsc *trieStatisticsCollector) Add(trieStats *TrieStatsDTO) {
	tsc.numNodes += trieStats.TotalNumNodes
	tsc.triesSize += trieStats.TotalNodesSize
	tsc.numDataTries++

	insertInSortedArray(tsc.triesBySize, trieStats, isLessSize, tsc.numTriesToPrint)
	insertInSortedArray(tsc.triesByDepth, trieStats, isLessDeep, tsc.numTriesToPrint)
}

func (tsc *trieStatisticsCollector) Print() {
	triesBySize := " \n top " + strconv.Itoa(tsc.numTriesToPrint) + " tries by size"
	triesByDepth := " \n top " + strconv.Itoa(tsc.numTriesToPrint) + " tries by depth"

	log.Debug("tries statistics",
		"num of nodes copied", tsc.numNodes,
		"total size copied", core.ConvertBytes(tsc.triesSize),
		"num data tries copied", tsc.numDataTries,
		triesBySize, getOrderedTries(tsc.triesBySize),
		triesByDepth, getOrderedTries(tsc.triesByDepth),
	)
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
	numTriesToPrint int,
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
