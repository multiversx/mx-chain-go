package statistics

import (
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core"
)

type trieStatistics struct {
	address  string
	rootHash []byte

	maxTrieDepth   uint32
	branchNodes    *nodesStatistics
	extensionNodes *nodesStatistics
	leafNodes      *nodesStatistics
	migrationStats map[core.TrieNodeVersion]uint64

	mutex sync.RWMutex
}

type nodesStatistics struct {
	nodesSize uint64
	numNodes  uint64
}

// NewTrieStatistics creates a new instance of trieStatistics
func NewTrieStatistics() *trieStatistics {
	return &trieStatistics{
		address:      "",
		rootHash:     nil,
		maxTrieDepth: 0,
		branchNodes: &nodesStatistics{
			nodesSize: 0,
			numNodes:  0,
		},
		extensionNodes: &nodesStatistics{
			nodesSize: 0,
			numNodes:  0,
		},
		leafNodes: &nodesStatistics{
			nodesSize: 0,
			numNodes:  0,
		},
		migrationStats: make(map[core.TrieNodeVersion]uint64),
	}
}

// AddBranchNode will add the given level and size to the branch nodes statistics
func (ts *trieStatistics) AddBranchNode(level int, size uint64) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	ts.collectNodeStatistics(level, size, ts.branchNodes)
}

// AddExtensionNode will add the given level and size to the extension nodes statistics
func (ts *trieStatistics) AddExtensionNode(level int, size uint64) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	ts.collectNodeStatistics(level, size, ts.extensionNodes)
}

// AddLeafNode will add the given level and size to the leaf nodes statistics
func (ts *trieStatistics) AddLeafNode(level int, size uint64, version core.TrieNodeVersion) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	ts.collectNodeStatistics(level, size, ts.leafNodes)
	ts.migrationStats[version]++
}

// AddAccountInfo will add the address and rootHash to  the collected statistics
func (ts *trieStatistics) AddAccountInfo(address string, rootHash []byte) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	ts.address = address
	ts.rootHash = rootHash
}

func (ts *trieStatistics) collectNodeStatistics(level int, size uint64, nodeStats *nodesStatistics) {
	nodeStats.numNodes++
	nodeStats.nodesSize += size

	if uint32(level) > ts.maxTrieDepth {
		ts.maxTrieDepth = uint32(level)
	}
}

func (ts *trieStatistics) GetTotalNodesSize() uint64 {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	return ts.getTotalNodesSize()
}

func (ts *trieStatistics) getTotalNodesSize() uint64 {
	return ts.branchNodes.nodesSize + ts.extensionNodes.nodesSize + ts.leafNodes.nodesSize
}

func (ts *trieStatistics) GetTotalNumNodes() uint64 {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	return ts.getTotalNumNodes()
}

func (ts *trieStatistics) getTotalNumNodes() uint64 {
	return ts.branchNodes.numNodes + ts.extensionNodes.numNodes + ts.leafNodes.numNodes
}

func (ts *trieStatistics) GetMaxTrieDepth() uint32 {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	return ts.maxTrieDepth
}

func (ts *trieStatistics) GetBranchNodesSize() uint64 {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	return ts.branchNodes.nodesSize
}

func (ts *trieStatistics) GetNumBranchNodes() uint64 {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	return ts.branchNodes.numNodes
}

func (ts *trieStatistics) GetExtensionNodesSize() uint64 {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	return ts.extensionNodes.nodesSize
}

func (ts *trieStatistics) GetNumExtensionNodes() uint64 {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	return ts.extensionNodes.numNodes
}

func (ts *trieStatistics) GetLeafNodesSize() uint64 {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	return ts.leafNodes.nodesSize
}

func (ts *trieStatistics) GetNumLeafNodes() uint64 {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	return ts.leafNodes.numNodes
}

func (ts *trieStatistics) GetLeavesMigrationStats() map[core.TrieNodeVersion]uint64 {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	return ts.migrationStats
}

// IsInterfaceNil returns true if there is no value under the interface
func (ts *trieStatistics) IsInterfaceNil() bool {
	return ts == nil
}

// ToString returns the collected statistics as a string array
func (ts *trieStatistics) ToString() []string {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	stats := make([]string, 0)
	stats = append(stats, fmt.Sprintf("address %v,", ts.address))
	stats = append(stats, fmt.Sprintf("rootHash %v,", hex.EncodeToString(ts.rootHash)))
	stats = append(stats, fmt.Sprintf("total trie size = %v,", core.ConvertBytes(ts.getTotalNodesSize())))
	stats = append(stats, fmt.Sprintf("num trie nodes =  %v,", ts.getTotalNumNodes()))
	stats = append(stats, fmt.Sprintf("max trie depth = %v,", ts.maxTrieDepth))
	stats = append(stats, fmt.Sprintf("branch nodes size %v,", core.ConvertBytes(ts.branchNodes.nodesSize)))
	stats = append(stats, fmt.Sprintf("extension nodes size %v,", core.ConvertBytes(ts.extensionNodes.nodesSize)))
	stats = append(stats, fmt.Sprintf("leaf nodes size %v,", core.ConvertBytes(ts.leafNodes.nodesSize)))
	stats = append(stats, fmt.Sprintf("num branches %v,", ts.branchNodes.numNodes))
	stats = append(stats, fmt.Sprintf("num extensions %v,", ts.extensionNodes.numNodes))
	stats = append(stats, fmt.Sprintf("num leaves %v", ts.leafNodes.numNodes))
	stats = append(stats, getMigrationStatsString(ts.migrationStats)...)
	return stats
}

func getMigrationStatsString(migrationStats map[core.TrieNodeVersion]uint64) []string {
	stats := make([]string, 0)
	for version, numNodes := range migrationStats {
		stats = append(stats, fmt.Sprintf("num leaves with %s version = %v", core.GetStringForVersion(version), numNodes))
	}

	return stats
}
