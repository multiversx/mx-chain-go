package state

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/trie/statistics"
)

const numTriesToPrint = 10

type snapshotStatistics struct {
	numNodes     uint64
	numDataTries uint64
	triesSize    uint64
	triesBySize  []*statistics.TrieStatsDTO
	triesByDepth []*statistics.TrieStatsDTO

	startTime time.Time

	wgSnapshot *sync.WaitGroup
	wgSync     *sync.WaitGroup
	mutex      sync.RWMutex
}

func newSnapshotStatistics(snapshotDelta int, syncDelta int) *snapshotStatistics {
	wgSnapshot := &sync.WaitGroup{}
	wgSnapshot.Add(snapshotDelta)

	wgSync := &sync.WaitGroup{}
	wgSync.Add(syncDelta)
	return &snapshotStatistics{
		wgSnapshot:   wgSnapshot,
		wgSync:       wgSync,
		startTime:    time.Now(),
		triesBySize:  make([]*statistics.TrieStatsDTO, numTriesToPrint),
		triesByDepth: make([]*statistics.TrieStatsDTO, numTriesToPrint),
	}
}

// AddSize will add the given size to the trie size counter
func (ss *snapshotStatistics) AddSize(size uint64) {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	ss.numNodes++
	ss.triesSize += size
}

// SnapshotFinished marks the ending of a snapshot goroutine
func (ss *snapshotStatistics) SnapshotFinished() {
	ss.wgSnapshot.Done()
}

// NewSnapshotStarted marks the starting of a new snapshot goroutine
func (ss *snapshotStatistics) NewSnapshotStarted() {
	ss.wgSnapshot.Add(1)
}

// NewDataTrie increases the data Tries counter
func (ss *snapshotStatistics) NewDataTrie() {
	ss.mutex.Lock()
	defer ss.mutex.Unlock()

	ss.numDataTries++
}

// WaitForSnapshotsToFinish will wait until the waitGroup counter is zero
func (ss *snapshotStatistics) WaitForSnapshotsToFinish() {
	ss.wgSnapshot.Wait()
}

// AddTrieStats adds the given trie stats to the snapshot statistics
func (ss *snapshotStatistics) AddTrieStats(trieStats common.TrieStatisticsHandler) {
	ts := trieStats.GetTrieStats()

	ss.numNodes += ts.TotalNumNodes
	ss.triesSize += ts.TotalNodesSize
	ss.numDataTries++

	insertInSortedArray(ss.triesBySize, ts, isLessSize)
	insertInSortedArray(ss.triesByDepth, ts, isLessDeep)
}

func isLessSize(a *statistics.TrieStatsDTO, b *statistics.TrieStatsDTO) bool {
	return a.TotalNodesSize < b.TotalNodesSize
}

func isLessDeep(a *statistics.TrieStatsDTO, b *statistics.TrieStatsDTO) bool {
	return a.MaxTrieDepth < b.MaxTrieDepth
}

func insertInSortedArray(
	array []*statistics.TrieStatsDTO,
	ts *statistics.TrieStatsDTO,
	isLess func(*statistics.TrieStatsDTO, *statistics.TrieStatsDTO) bool,
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

// WaitForSyncToFinish will wait until the waitGroup counter is zero
func (ss *snapshotStatistics) WaitForSyncToFinish() {
	ss.wgSync.Wait()
}

// SyncFinished marks the end of the sync process
func (ss *snapshotStatistics) SyncFinished() {
	ss.wgSync.Done()
}

// PrintStats will print the stats after the snapshot has finished
func (ss *snapshotStatistics) PrintStats(identifier string, rootHash []byte) {
	ss.mutex.RLock()
	defer ss.mutex.RUnlock()

	triesBySize := " \n top " + strconv.Itoa(numTriesToPrint) + " tries by size"
	triesByDepth := " \n top " + strconv.Itoa(numTriesToPrint) + " tries by depth"
	log.Debug("snapshot statistics",
		"type", identifier,
		"duration", time.Since(ss.startTime).Truncate(time.Second),
		"num of nodes copied", ss.numNodes,
		"total size copied", core.ConvertBytes(ss.triesSize),
		"num data tries copied", ss.numDataTries,
		"rootHash", rootHash,
		triesBySize, getOrderedTries(ss.triesBySize),
		triesByDepth, getOrderedTries(ss.triesByDepth),
	)
}

func getOrderedTries(tries []*statistics.TrieStatsDTO) string {
	triesStats := make([]string, 0)
	for i := 0; i < len(tries); i++ {
		if tries[i] == nil {
			continue
		}
		triesStats = append(triesStats, strings.Join(tries[i].ToString(), " "))
	}

	return strings.Join(triesStats, "\n")
}
