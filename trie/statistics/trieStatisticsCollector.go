package statistics

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("trieStatistics")

const numTriesToPrint = 10

type trieStatisticsCollector struct {
	trieStatsByType map[common.TrieType]common.TrieStatisticsHandler
	triesBySize     []common.TrieStatisticsHandler
	triesByDepth    []common.TrieStatisticsHandler
	numTriesByType  map[common.TrieType]uint64

	mutex sync.RWMutex
}

// NewTrieStatisticsCollector creates a new instance of trieStatisticsCollector
func NewTrieStatisticsCollector() *trieStatisticsCollector {
	return &trieStatisticsCollector{
		trieStatsByType: make(map[common.TrieType]common.TrieStatisticsHandler),
		triesBySize:     make([]common.TrieStatisticsHandler, numTriesToPrint),
		triesByDepth:    make([]common.TrieStatisticsHandler, numTriesToPrint),
		numTriesByType:  make(map[common.TrieType]uint64),
	}
}

// Add adds the given trie statistics to the statistics collector
func (tsc *trieStatisticsCollector) Add(trieStats common.TrieStatisticsHandler, trieType common.TrieType) {
	if check.IfNil(trieStats) {
		log.Warn("programming error, nil trie stats received")
		return
	}

	tsc.mutex.Lock()
	defer tsc.mutex.Unlock()

	_, ok := tsc.trieStatsByType[trieType]
	if !ok {
		tsc.trieStatsByType[trieType] = NewTrieStatistics()
	}

	tsc.trieStatsByType[trieType].MergeTriesStatistics(trieStats)
	tsc.numTriesByType[trieType]++

	insertInSortedArray(tsc.triesBySize, trieStats, isLessSize)
	insertInSortedArray(tsc.triesByDepth, trieStats, isLessDeep)
}

// Print will print all the collected statistics
func (tsc *trieStatisticsCollector) Print() {
	tsc.mutex.RLock()
	defer tsc.mutex.RUnlock()

	triesBySize := " \n top " + strconv.Itoa(numTriesToPrint) + " tries by size \n"
	triesByDepth := " \n top " + strconv.Itoa(numTriesToPrint) + " tries by depth \n"

	totalNumNodes := uint64(0)
	totalStateSize := uint64(0)
	numMainTrieLeaves := uint64(0)
	maxDepthMainTrie := uint32(0)

	for trieType, stats := range tsc.trieStatsByType {
		totalNumNodes += stats.GetTotalNumNodes()
		totalStateSize += stats.GetTotalNodesSize()

		if trieType == common.MainTrie {
			numMainTrieLeaves = stats.GetNumLeafNodes()
			maxDepthMainTrie = stats.GetMaxTrieDepth()
		}
	}

	log.Debug("tries statistics",
		"num of nodes", totalNumNodes,
		"total size", core.ConvertBytes(totalStateSize),
		"num tries by type", getNumTriesByTypeString(tsc.numTriesByType),
		"num main trie leaves", numMainTrieLeaves,
		"max depth main trie", maxDepthMainTrie,

		triesBySize, getOrderedTries(tsc.triesBySize),
		triesByDepth, getOrderedTries(tsc.triesByDepth),
	)

	for trieType, trieStats := range tsc.trieStatsByType {
		message := fmt.Sprintf("migration stats for %v", trieType)
		log.Debug(message, "stats", getMigrationStatsString(trieStats.GetLeavesMigrationStats()))
	}

	if log.GetLevel() == logger.LogTrace {
		tsc.printDetailedTriesStatistics()
	}
}

func getNumTriesByTypeString(numTriesByTypeMap map[common.TrieType]uint64) string {
	var numTriesByTypeMapString []string
	for trieType, numTries := range numTriesByTypeMap {
		numTriesByTypeMapString = append(numTriesByTypeMapString, fmt.Sprintf("%v: %v", trieType, numTries))
	}

	sort.Slice(numTriesByTypeMapString, func(i, j int) bool {
		return numTriesByTypeMapString[i] < numTriesByTypeMapString[j]
	})

	return strings.Join(numTriesByTypeMapString, ", ")
}

func (tsc *trieStatisticsCollector) printDetailedTriesStatistics() {
	totalNumBranches := uint64(0)
	totalNumExtensions := uint64(0)
	totalNumLeaves := uint64(0)
	totalSizeBranches := uint64(0)
	totalSizeExtensions := uint64(0)
	totalSizeLeaves := uint64(0)

	for _, stats := range tsc.trieStatsByType {
		totalNumBranches += stats.GetNumBranchNodes()
		totalNumExtensions += stats.GetNumExtensionNodes()
		totalNumLeaves += stats.GetNumLeafNodes()
		totalSizeBranches += stats.GetBranchNodesSize()
		totalSizeExtensions += stats.GetExtensionNodesSize()
		totalSizeLeaves += stats.GetLeafNodesSize()
	}

	log.Trace("detailed tries statistics",
		"total num branches", totalNumBranches,
		"total num extensions", totalNumExtensions,
		"total num leaves", totalNumLeaves,
		"total size branches", core.ConvertBytes(totalSizeBranches),
		"total size extensions", core.ConvertBytes(totalSizeExtensions),
		"total size leaves", core.ConvertBytes(totalSizeLeaves),
	)
}

// GetNumNodes returns the number of nodes
func (tsc *trieStatisticsCollector) GetNumNodes() uint64 {
	tsc.mutex.RLock()
	defer tsc.mutex.RUnlock()

	totalNumNodes := uint64(0)
	for _, stats := range tsc.trieStatsByType {
		totalNumNodes += stats.GetTotalNumNodes()
	}

	return totalNumNodes
}

func getOrderedTries(tries []common.TrieStatisticsHandler) string {
	triesStats := make([]string, 0)
	for i := 0; i < len(tries); i++ {
		if check.IfNil(tries[i]) {
			continue
		}
		triesStats = append(triesStats, strings.Join(tries[i].ToString(), " "))
	}

	return strings.Join(triesStats, "\n")
}

func isLessSize(a common.TrieStatisticsHandler, b common.TrieStatisticsHandler) bool {
	return a.GetTotalNodesSize() < b.GetTotalNodesSize()
}

func isLessDeep(a common.TrieStatisticsHandler, b common.TrieStatisticsHandler) bool {
	return a.GetMaxTrieDepth() < b.GetMaxTrieDepth()
}

func insertInSortedArray(
	array []common.TrieStatisticsHandler,
	ts common.TrieStatisticsHandler,
	isLess func(common.TrieStatisticsHandler, common.TrieStatisticsHandler) bool,
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
