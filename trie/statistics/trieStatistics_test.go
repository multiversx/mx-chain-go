package statistics

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrieStatistics_AddBranchNode(t *testing.T) {
	t.Parallel()

	ts := NewTrieStatistics()

	level := 2
	size := uint64(15)
	ts.AddBranchNode(level, size)
	ts.AddBranchNode(level+1, size)
	ts.AddBranchNode(level-1, size)
	assert.Equal(t, uint64(3), ts.branchNodes.numNodes)
	assert.Equal(t, 3*size, ts.branchNodes.nodesSize)
	assert.Equal(t, 3*size, ts.branchNodes.nodesSize)
	assert.Equal(t, uint32(level+1), ts.maxTrieDepth)
}

func TestTrieStatistics_AddExtensionNode(t *testing.T) {
	t.Parallel()

	ts := NewTrieStatistics()

	level := 2
	size := uint64(15)
	ts.AddExtensionNode(level, size)
	ts.AddExtensionNode(level+1, size)
	ts.AddExtensionNode(level-1, size)
	assert.Equal(t, uint64(3), ts.extensionNodes.numNodes)
	assert.Equal(t, 3*size, ts.extensionNodes.nodesSize)
	assert.Equal(t, 3*size, ts.extensionNodes.nodesSize)
	assert.Equal(t, uint32(level+1), ts.maxTrieDepth)
}

func TestTrieStatistics_AddLeafNode(t *testing.T) {
	t.Parallel()

	ts := NewTrieStatistics()

	level := 2
	size := uint64(15)
	ts.AddLeafNode(level, size)
	ts.AddLeafNode(level+1, size)
	ts.AddLeafNode(level-1, size)
	assert.Equal(t, uint64(3), ts.leafNodes.numNodes)
	assert.Equal(t, 3*size, ts.leafNodes.nodesSize)
	assert.Equal(t, 3*size, ts.leafNodes.nodesSize)
	assert.Equal(t, uint32(level+1), ts.maxTrieDepth)
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
	numBranches := 24
	expectedBranchesSize := numBranches * branchSize

	extensionSize := 5
	numExtensions := 3
	expectedExtensionsSize := extensionSize * numExtensions

	leafSize := 50
	numLeaves := 20
	expectedLeavesSize := leafSize * numLeaves

	totalNumNodes := uint64(47)

	for i := 0; i < numBranches; i++ {
		ts.AddBranchNode(i, uint64(branchSize))
	}
	for i := 0; i < numExtensions; i++ {
		ts.AddExtensionNode(i, uint64(extensionSize))
	}
	for i := 0; i < numLeaves; i++ {
		ts.AddLeafNode(i, uint64(leafSize))
	}

	stats := ts.GetTrieStats()
	assert.Equal(t, uint32(numBranches-1), stats.MaxTrieDepth)
	assert.Equal(t, uint64(expectedBranchesSize), stats.BranchNodesSize)
	assert.Equal(t, uint64(expectedExtensionsSize), stats.ExtensionNodesSize)
	assert.Equal(t, uint64(expectedLeavesSize), stats.LeafNodesSize)
	assert.Equal(t, uint64(expectedLeavesSize+expectedBranchesSize+expectedExtensionsSize), stats.TotalNodesSize)
	assert.Equal(t, totalNumNodes, stats.TotalNumNodes)
	assert.Equal(t, uint64(numBranches), stats.NumBranchNodes)
	assert.Equal(t, uint64(numExtensions), stats.NumExtensionNodes)
	assert.Equal(t, uint64(numLeaves), stats.NumLeafNodes)
}

func TestTrieStatsDTO_ToString(t *testing.T) {
	t.Parallel()

	tsd := TrieStatsDTO{
		Address:            "address",
		RootHash:           []byte("root hash"),
		TotalNodesSize:     1,
		TotalNumNodes:      1,
		MaxTrieDepth:       1,
		BranchNodesSize:    1,
		NumBranchNodes:     1,
		ExtensionNodesSize: 1,
		NumExtensionNodes:  1,
		LeafNodesSize:      1,
		NumLeafNodes:       1,
	}

	expectedLines := []string{
		fmt.Sprintf("address %v,", tsd.Address),
		fmt.Sprintf("rootHash %v,", hex.EncodeToString(tsd.RootHash)),
		fmt.Sprintf("total trie size = %v,", core.ConvertBytes(tsd.TotalNodesSize)),
		fmt.Sprintf("num trie nodes =  %v,", tsd.TotalNumNodes),
		fmt.Sprintf("max trie depth = %v,", tsd.MaxTrieDepth),
		fmt.Sprintf("branch nodes size %v,", core.ConvertBytes(tsd.BranchNodesSize)),
		fmt.Sprintf("extension nodes size %v,", core.ConvertBytes(tsd.ExtensionNodesSize)),
		fmt.Sprintf("leaf nodes size %v,", core.ConvertBytes(tsd.LeafNodesSize)),
		fmt.Sprintf("num branches %v,", tsd.NumBranchNodes),
		fmt.Sprintf("num extensions %v,", tsd.NumExtensionNodes),
		fmt.Sprintf("num leaves %v", tsd.NumLeafNodes),
	}
	stringDTO := tsd.ToString()
	for i, line := range stringDTO {
		require.Equal(t, expectedLines[i], line)
	}
}
