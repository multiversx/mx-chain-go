package statistics

import (
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
)

type trieStatistics struct {
	address  string
	rootHash []byte

	maxTrieDepth   uint32
	branchNodes    *nodesStatistics
	extensionNodes *nodesStatistics
	leafNodes      *nodesStatistics
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
	}
}

// AddBranchNode will add the given level and size to the branch nodes statistics
func (ts *trieStatistics) AddBranchNode(level int, size uint64) {
	ts.collectNodeStatistics(level, size, ts.branchNodes)
}

// AddExtensionNode will add the given level and size to the extension nodes statistics
func (ts *trieStatistics) AddExtensionNode(level int, size uint64) {
	ts.collectNodeStatistics(level, size, ts.extensionNodes)
}

// AddLeafNode will add the given level and size to the leaf nodes statistics
func (ts *trieStatistics) AddLeafNode(level int, size uint64) {
	ts.collectNodeStatistics(level, size, ts.leafNodes)
}

// AddAccountInfo will add the address and rootHash to  the collected statistics
func (ts *trieStatistics) AddAccountInfo(address string, rootHash []byte) {
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

// GetTrieStats returns a DTO that contains all the collected info about the trie
func (ts *trieStatistics) GetTrieStats() *TrieStatsDTO {
	totalNodesSize := ts.branchNodes.nodesSize + ts.extensionNodes.nodesSize + ts.leafNodes.nodesSize
	totalNumNodes := ts.branchNodes.numNodes + ts.extensionNodes.numNodes + ts.leafNodes.numNodes

	return &TrieStatsDTO{
		Address:            ts.address,
		RootHash:           ts.rootHash,
		TotalNodesSize:     totalNodesSize,
		TotalNumNodes:      totalNumNodes,
		MaxTrieDepth:       ts.maxTrieDepth,
		BranchNodesSize:    ts.branchNodes.nodesSize,
		NumBranchNodes:     ts.branchNodes.numNodes,
		ExtensionNodesSize: ts.extensionNodes.nodesSize,
		NumExtensionNodes:  ts.extensionNodes.numNodes,
		LeafNodesSize:      ts.leafNodes.nodesSize,
		NumLeafNodes:       ts.leafNodes.numNodes,
	}
}

// TrieStatsDTO holds the statistics for the trie
type TrieStatsDTO struct {
	Address        string
	RootHash       []byte
	TotalNodesSize uint64
	TotalNumNodes  uint64
	MaxTrieDepth   uint32

	BranchNodesSize    uint64
	NumBranchNodes     uint64
	ExtensionNodesSize uint64
	NumExtensionNodes  uint64
	LeafNodesSize      uint64
	NumLeafNodes       uint64
}

// ToString returns the collected statistics as a string array
func (tsd *TrieStatsDTO) ToString() []string {
	stats := make([]string, 0)
	stats = append(stats, fmt.Sprintf("address %v,", tsd.Address))
	stats = append(stats, fmt.Sprintf("rootHash %v,", hex.EncodeToString(tsd.RootHash)))
	stats = append(stats, fmt.Sprintf("total trie size = %v,", core.ConvertBytes(tsd.TotalNodesSize)))
	stats = append(stats, fmt.Sprintf("num trie nodes =  %v,", tsd.TotalNumNodes))
	stats = append(stats, fmt.Sprintf("max trie depth = %v,", tsd.MaxTrieDepth))
	stats = append(stats, fmt.Sprintf("branch nodes size %v,", core.ConvertBytes(tsd.BranchNodesSize)))
	stats = append(stats, fmt.Sprintf("extension nodes size %v,", core.ConvertBytes(tsd.ExtensionNodesSize)))
	stats = append(stats, fmt.Sprintf("leaf nodes size %v,", core.ConvertBytes(tsd.LeafNodesSize)))
	stats = append(stats, fmt.Sprintf("num branches  %v,", tsd.NumBranchNodes))
	stats = append(stats, fmt.Sprintf("num extensions  %v,", tsd.NumExtensionNodes))
	stats = append(stats, fmt.Sprintf("num leaves  %v", tsd.NumLeafNodes))
	return stats
}
