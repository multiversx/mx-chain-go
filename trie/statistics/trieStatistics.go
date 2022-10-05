package statistics

import (
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
)

type trieStatistics struct {
	address  []byte
	rootHash []byte

	branchNodes    *nodesStatistics
	extensionNodes *nodesStatistics
	leafNodes      *nodesStatistics
}

type nodesStatistics struct {
	nodesSize     uint64
	nodesPerLevel []uint32
}

func NewTrieStatistics() *trieStatistics {
	return &trieStatistics{
		address:  nil,
		rootHash: nil,
		branchNodes: &nodesStatistics{
			nodesSize:     0,
			nodesPerLevel: make([]uint32, 0),
		},
		extensionNodes: &nodesStatistics{
			nodesSize:     0,
			nodesPerLevel: make([]uint32, 0),
		},
		leafNodes: &nodesStatistics{
			nodesSize:     0,
			nodesPerLevel: make([]uint32, 0),
		},
	}
}

func (ts *trieStatistics) AddBranchNode(level int, size uint64) {
	collectNodeStatistics(level, size, ts.branchNodes)
}

func (ts *trieStatistics) AddExtensionNode(level int, size uint64) {
	collectNodeStatistics(level, size, ts.extensionNodes)
}

func (ts *trieStatistics) AddLeafNode(level int, size uint64) {
	collectNodeStatistics(level, size, ts.leafNodes)
}

func (ts *trieStatistics) AddAccountInfo(address []byte, rootHash []byte) {
	ts.address = address
	ts.rootHash = rootHash
}

func collectNodeStatistics(level int, size uint64, nodeStats *nodesStatistics) {
	numLevels := len(nodeStats.nodesPerLevel)
	for i := numLevels; i <= level; i++ {
		nodeStats.nodesPerLevel = append(nodeStats.nodesPerLevel, 0)
	}

	nodeStats.nodesPerLevel[level]++
	nodeStats.nodesSize += size
}

func (ts *trieStatistics) GetTrieStats() *TrieStatsDTO {
	numLevelsWithBranches := len(ts.branchNodes.nodesPerLevel)
	numLevelsWithExtensions := len(ts.extensionNodes.nodesPerLevel)
	numLevelsWithLeaves := len(ts.leafNodes.nodesPerLevel)

	maxTrieDepth := numLevelsWithBranches
	if numLevelsWithExtensions > maxTrieDepth {
		maxTrieDepth = numLevelsWithExtensions
	}
	if numLevelsWithLeaves > maxTrieDepth {
		maxTrieDepth = numLevelsWithLeaves
	}

	totalNumNodesPerLevel := make([]uint32, 0)
	totalNumNodes := uint64(0)
	for i := 0; i < maxTrieDepth; i++ {
		nodesPerLevel := uint32(0)

		if i < numLevelsWithBranches {
			nodesPerLevel += ts.branchNodes.nodesPerLevel[i]
		}
		if i < numLevelsWithExtensions {
			nodesPerLevel += ts.extensionNodes.nodesPerLevel[i]
		}
		if i < numLevelsWithLeaves {
			nodesPerLevel += ts.leafNodes.nodesPerLevel[i]
		}

		totalNumNodesPerLevel = append(totalNumNodesPerLevel, nodesPerLevel)
		totalNumNodes += uint64(nodesPerLevel)
	}

	totalNodesSize := ts.branchNodes.nodesSize + ts.extensionNodes.nodesSize + ts.leafNodes.nodesSize

	return &TrieStatsDTO{
		Address:               ts.address,
		RootHash:              ts.rootHash,
		NumBranchesPerLevel:   ts.branchNodes.nodesPerLevel,
		NumExtensionsPerLevel: ts.extensionNodes.nodesPerLevel,
		NumLeavesPerLevel:     ts.leafNodes.nodesPerLevel,
		TotalNumNodesPerLevel: totalNumNodesPerLevel,
		TotalNumNodes:         totalNumNodes,
		BranchNodesSize:       ts.branchNodes.nodesSize,
		ExtensionNodesSize:    ts.extensionNodes.nodesSize,
		LeafNodesSize:         ts.leafNodes.nodesSize,
		TotalNodesSize:        totalNodesSize,
		MaxTrieDepth:          uint32(maxTrieDepth),
	}
}

type TrieStatsDTO struct {
	Address        []byte
	RootHash       []byte
	TotalNodesSize uint64
	TotalNumNodes  uint64
	MaxTrieDepth   uint32

	TotalNumNodesPerLevel []uint32
	NumBranchesPerLevel   []uint32
	NumExtensionsPerLevel []uint32
	NumLeavesPerLevel     []uint32

	BranchNodesSize    uint64
	ExtensionNodesSize uint64
	LeafNodesSize      uint64
}

func (tsd *TrieStatsDTO) ToString() []string {
	stats := make([]string, 0)
	stats = append(stats, fmt.Sprintf("address %v,", hex.EncodeToString(tsd.Address)))
	stats = append(stats, fmt.Sprintf("rootHash %v,", hex.EncodeToString(tsd.RootHash)))
	stats = append(stats, fmt.Sprintf("total trie size = %v,", core.ConvertBytes(tsd.TotalNodesSize)))
	stats = append(stats, fmt.Sprintf("num trie nodes =  %v,", tsd.TotalNumNodes))
	stats = append(stats, fmt.Sprintf("max trie depth = %v,", tsd.MaxTrieDepth))
	stats = append(stats, fmt.Sprintf("branch nodes size %v,", core.ConvertBytes(tsd.BranchNodesSize)))
	stats = append(stats, fmt.Sprintf("extension nodes size %v,", core.ConvertBytes(tsd.ExtensionNodesSize)))
	stats = append(stats, fmt.Sprintf("leaf nodes size %v,", core.ConvertBytes(tsd.LeafNodesSize)))
	stats = append(stats, fmt.Sprintf("total nodes per level %v,", tsd.TotalNumNodesPerLevel))
	stats = append(stats, fmt.Sprintf("num branches per level %v,", tsd.NumBranchesPerLevel))
	stats = append(stats, fmt.Sprintf("num extensions per level %v,", tsd.NumExtensionsPerLevel))
	stats = append(stats, fmt.Sprintf("num leaves per level %v", tsd.NumLeavesPerLevel))
	return stats
}
