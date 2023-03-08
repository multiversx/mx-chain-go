package statistics

import (
	"strconv"
	"strings"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/common"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("trieStatistics")

const numTriesToPrint = 10

type trieStatisticsCollector struct {
	numNodes            uint64
	numDataTries        uint64
	triesSize           uint64
	numTotalLeaves      uint64
	numTotalExtensions  uint64
	numTotalBranches    uint64
	totalSizeLeaves     uint64
	totalSizeExtensions uint64
	totalSizeBranches   uint64
	triesBySize         []common.TrieStatisticsHandler
	triesByDepth        []common.TrieStatisticsHandler
	migrationStats      map[core.TrieNodeVersion]uint64

	mutex sync.RWMutex
}

// NewTrieStatisticsCollector creates a new instance of trieStatisticsCollector
func NewTrieStatisticsCollector() *trieStatisticsCollector {
	return &trieStatisticsCollector{
		numNodes:            0,
		numDataTries:        0,
		triesSize:           0,
		numTotalLeaves:      0,
		numTotalExtensions:  0,
		numTotalBranches:    0,
		totalSizeLeaves:     0,
		totalSizeExtensions: 0,
		totalSizeBranches:   0,
		triesBySize:         make([]common.TrieStatisticsHandler, numTriesToPrint),
		triesByDepth:        make([]common.TrieStatisticsHandler, numTriesToPrint),
		migrationStats:      make(map[core.TrieNodeVersion]uint64),
	}
}

// Add adds the given trie statistics to the statistics collector
func (tsc *trieStatisticsCollector) Add(trieStats common.TrieStatisticsHandler) {
	if trieStats == nil {
		log.Warn("programming error, nil trie stats received")
		return
	}

	tsc.mutex.Lock()
	defer tsc.mutex.Unlock()

	tsc.numNodes += trieStats.GetTotalNumNodes()
	tsc.triesSize += trieStats.GetTotalNodesSize()
	tsc.numDataTries++

	tsc.numTotalBranches += trieStats.GetNumBranchNodes()
	tsc.numTotalExtensions += trieStats.GetNumExtensionNodes()
	tsc.numTotalLeaves += trieStats.GetNumLeafNodes()
	tsc.totalSizeBranches += trieStats.GetBranchNodesSize()
	tsc.totalSizeExtensions += trieStats.GetExtensionNodesSize()
	tsc.totalSizeLeaves += trieStats.GetLeafNodesSize()
	for version, numNodes := range trieStats.GetLeavesMigrationStats() {
		tsc.migrationStats[version] += numNodes
	}

	insertInSortedArray(tsc.triesBySize, trieStats, isLessSize)
	insertInSortedArray(tsc.triesByDepth, trieStats, isLessDeep)
}

// Print will print all the collected statistics
func (tsc *trieStatisticsCollector) Print() {
	tsc.mutex.RLock()
	defer tsc.mutex.RUnlock()

	triesBySize := " \n top " + strconv.Itoa(numTriesToPrint) + " tries by size \n"
	triesByDepth := " \n top " + strconv.Itoa(numTriesToPrint) + " tries by depth \n"

	log.Debug("tries statistics",
		"num of nodes", tsc.numNodes,
		"total size", core.ConvertBytes(tsc.triesSize),
		"num tries", tsc.numDataTries,
		"total num branches", tsc.numTotalBranches,
		"total num extensions", tsc.numTotalExtensions,
		"total num leaves", tsc.numTotalLeaves,
		"total size branches", core.ConvertBytes(tsc.totalSizeBranches),
		"total size extensions", core.ConvertBytes(tsc.totalSizeExtensions),
		"total size leaves", core.ConvertBytes(tsc.totalSizeLeaves),
		"migration stats", getMigrationStatsString(tsc.migrationStats),
		triesBySize, getOrderedTries(tsc.triesBySize),
		triesByDepth, getOrderedTries(tsc.triesByDepth),
	)
}

// GetNumNodes returns the number of nodes
func (tsc *trieStatisticsCollector) GetNumNodes() uint64 {
	tsc.mutex.RLock()
	defer tsc.mutex.RUnlock()
	
	return tsc.numNodes
}

func getOrderedTries(tries []common.TrieStatisticsHandler) string {
	triesStats := make([]string, 0)
	for i := 0; i < len(tries); i++ {
		if tries[i] == nil {
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
