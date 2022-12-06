package statistics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrieStatistics_AddBranchNode(t *testing.T) {
	t.Parallel()

	ts := NewTrieStatistics()

	level := 2
	size := uint64(15)
	ts.AddBranchNode(level, size)
	ts.AddBranchNode(level, size)
	assert.Equal(t, 3, len(ts.branchNodes.nodesPerLevel))
	assert.Equal(t, uint32(2), ts.branchNodes.nodesPerLevel[level])
	assert.Equal(t, 2*size, ts.branchNodes.nodesSize)
}

func TestTrieStatistics_AddExtensionNode(t *testing.T) {
	t.Parallel()

	ts := NewTrieStatistics()

	level := 2
	size := uint64(15)
	ts.AddExtensionNode(level, size)
	ts.AddExtensionNode(level, size)
	assert.Equal(t, 3, len(ts.extensionNodes.nodesPerLevel))
	assert.Equal(t, uint32(2), ts.extensionNodes.nodesPerLevel[level])
	assert.Equal(t, 2*size, ts.extensionNodes.nodesSize)
}

func TestTrieStatistics_AddLeafNode(t *testing.T) {
	t.Parallel()

	ts := NewTrieStatistics()

	level := 2
	size := uint64(15)
	ts.AddLeafNode(level, size)
	ts.AddLeafNode(level, size)
	assert.Equal(t, 3, len(ts.leafNodes.nodesPerLevel))
	assert.Equal(t, uint32(2), ts.leafNodes.nodesPerLevel[level])
	assert.Equal(t, 2*size, ts.leafNodes.nodesSize)
}

func TestTrieStatistics_AddAccountInfo(t *testing.T) {
	t.Parallel()

	address := "address"
	rootHash := []byte("rootHash")

	ts := NewTrieStatistics()
	ts.AddAccountInfo(address, rootHash)

	assert.Equal(t, address, ts.address)
	assert.Equal(t, rootHash, ts.rootHash)
}

func TestTrieStatistics_GetTrieStats(t *testing.T) {
	t.Parallel()

	ts := NewTrieStatistics()

	branchSize := 30
	branchesPerLevel := []int{1, 6, 7, 10}
	numBranches := 24
	expectedBranchesSize := numBranches * branchSize

	extensionSize := 5
	extensionsPerLevel := []int{0, 1, 2}
	numExtensions := 3
	expectedExtensionsSize := extensionSize * numExtensions

	leafSize := 50
	leavesPerLevel := []int{0, 0, 2, 2, 16}
	numLeaves := 20
	expectedLeavesSize := leafSize * numLeaves

	totalNumNodesPerLevel := []uint32{1, 7, 11, 12, 16}
	totalNumNodes := uint64(47)

	for i, numBranchesPerLevel := range branchesPerLevel {
		for j := 0; j < numBranchesPerLevel; j++ {
			ts.AddBranchNode(i, uint64(branchSize))
		}
	}
	for i, numExtensionsPerLevel := range extensionsPerLevel {
		for j := 0; j < numExtensionsPerLevel; j++ {
			ts.AddExtensionNode(i, uint64(extensionSize))
		}
	}
	for i, numLeavesPerLevel := range leavesPerLevel {
		for j := 0; j < numLeavesPerLevel; j++ {
			ts.AddLeafNode(i, uint64(leafSize))
		}
	}

	stats := ts.GetTrieStats()
	assert.Equal(t, uint32(5), stats.MaxTrieDepth)
	assert.Equal(t, uint64(expectedBranchesSize), stats.BranchNodesSize)
	assert.Equal(t, uint64(expectedExtensionsSize), stats.ExtensionNodesSize)
	assert.Equal(t, uint64(expectedLeavesSize), stats.LeafNodesSize)
	assert.Equal(t, uint64(expectedLeavesSize+expectedBranchesSize+expectedExtensionsSize), stats.TotalNodesSize)
	assert.Equal(t, totalNumNodesPerLevel, stats.TotalNumNodesPerLevel)
	assert.Equal(t, totalNumNodes, stats.TotalNumNodes)
}
