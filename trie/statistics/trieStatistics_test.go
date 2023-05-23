package statistics

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/assert"
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
	ts.AddLeafNode(level, size, 0)
	ts.AddLeafNode(level+1, size, 0)
	ts.AddLeafNode(level-1, size, 1)
	assert.Equal(t, uint64(3), ts.leafNodes.numNodes)
	assert.Equal(t, 3*size, ts.leafNodes.nodesSize)
	assert.Equal(t, 3*size, ts.leafNodes.nodesSize)
	assert.Equal(t, uint32(level+1), ts.maxTrieDepth)
	assert.Equal(t, uint64(2), ts.migrationStats[0])
	assert.Equal(t, uint64(1), ts.migrationStats[1])
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
		ts.AddLeafNode(i, uint64(leafSize), 0)
	}

	assert.Equal(t, uint32(numBranches-1), ts.GetMaxTrieDepth())
	assert.Equal(t, uint64(expectedBranchesSize), ts.GetBranchNodesSize())
	assert.Equal(t, uint64(expectedExtensionsSize), ts.GetExtensionNodesSize())
	assert.Equal(t, uint64(expectedLeavesSize), ts.GetLeafNodesSize())
	assert.Equal(t, uint64(expectedLeavesSize+expectedBranchesSize+expectedExtensionsSize), ts.GetTotalNodesSize())
	assert.Equal(t, totalNumNodes, ts.GetTotalNumNodes())
	assert.Equal(t, uint64(numBranches), ts.GetNumBranchNodes())
	assert.Equal(t, uint64(numExtensions), ts.GetNumExtensionNodes())
	assert.Equal(t, uint64(numLeaves), ts.GetNumLeafNodes())
	assert.Equal(t, uint64(numLeaves), ts.GetLeavesMigrationStats()[0])
}

func TestTrieStatistics_MergeTriesStatistics(t *testing.T) {
	t.Parallel()

	leafSize := uint64(10)
	branchSize := uint64(8)
	extensionSize := uint64(5)

	ts := NewTrieStatistics()
	newTs := NewTrieStatistics()
	newTs.AddLeafNode(2, leafSize, 0)
	newTs.AddLeafNode(3, leafSize, 1)
	newTs.AddBranchNode(1, branchSize)
	newTs.AddExtensionNode(3, extensionSize)

	ts.MergeTriesStatistics(newTs)

	assert.Equal(t, uint32(3), ts.GetMaxTrieDepth())
	assert.Equal(t, branchSize, ts.GetBranchNodesSize())
	assert.Equal(t, extensionSize, ts.GetExtensionNodesSize())
	assert.Equal(t, 2*leafSize, ts.GetLeafNodesSize())
	assert.Equal(t, branchSize+extensionSize+2*leafSize, ts.GetTotalNodesSize())
	assert.Equal(t, uint64(4), ts.GetTotalNumNodes())
	assert.Equal(t, uint64(1), ts.GetNumBranchNodes())
	assert.Equal(t, uint64(1), ts.GetNumExtensionNodes())
	assert.Equal(t, uint64(2), ts.GetNumLeafNodes())
	assert.Equal(t, uint64(1), ts.GetLeavesMigrationStats()[0])
	assert.Equal(t, uint64(1), ts.GetLeavesMigrationStats()[1])

	newTs = NewTrieStatistics()
	newTs.AddLeafNode(4, leafSize, 0)
	newTs.AddLeafNode(1, leafSize, 1)
	newTs.AddBranchNode(1, branchSize)
	newTs.AddExtensionNode(3, extensionSize)

	ts.MergeTriesStatistics(newTs)
	totalNodesSize := branchSize*2 + extensionSize*2 + leafSize*4

	assert.Equal(t, uint32(4), ts.GetMaxTrieDepth())
	assert.Equal(t, branchSize*2, ts.GetBranchNodesSize())
	assert.Equal(t, extensionSize*2, ts.GetExtensionNodesSize())
	assert.Equal(t, leafSize*4, ts.GetLeafNodesSize())
	assert.Equal(t, totalNodesSize, ts.GetTotalNodesSize())
	assert.Equal(t, uint64(8), ts.GetTotalNumNodes())
	assert.Equal(t, uint64(2), ts.GetNumBranchNodes())
	assert.Equal(t, uint64(2), ts.GetNumExtensionNodes())
	assert.Equal(t, uint64(4), ts.GetNumLeafNodes())
	assert.Equal(t, uint64(2), ts.GetLeavesMigrationStats()[0])
	assert.Equal(t, uint64(2), ts.GetLeavesMigrationStats()[1])

	address := "address"
	rootHash := []byte("rootHash")
	ts.AddAccountInfo(address, rootHash)

	trieStatsStrings := ts.ToString()
	assert.Equal(t, 13, len(trieStatsStrings))
	assert.Equal(t, fmt.Sprintf("address %v,", address), trieStatsStrings[0])
	assert.Equal(t, fmt.Sprintf("rootHash %v,", hex.EncodeToString(rootHash)), trieStatsStrings[1])
	assert.Equal(t, fmt.Sprintf("total trie size = %v,", core.ConvertBytes(totalNodesSize)), trieStatsStrings[2])
	assert.Equal(t, fmt.Sprintf("num trie nodes =  %v,", 8), trieStatsStrings[3])
	assert.Equal(t, fmt.Sprintf("max trie depth = %v,", 4), trieStatsStrings[4])
	assert.Equal(t, fmt.Sprintf("branch nodes size %v,", core.ConvertBytes(16)), trieStatsStrings[5])
	assert.Equal(t, fmt.Sprintf("extension nodes size %v,", core.ConvertBytes(10)), trieStatsStrings[6])
	assert.Equal(t, fmt.Sprintf("leaf nodes size %v,", core.ConvertBytes(40)), trieStatsStrings[7])
	assert.Equal(t, fmt.Sprintf("num branches %v,", 2), trieStatsStrings[8])
	assert.Equal(t, fmt.Sprintf("num extensions %v,", 2), trieStatsStrings[9])
	assert.Equal(t, fmt.Sprintf("num leaves %v", 4), trieStatsStrings[10])
	assert.Equal(t, fmt.Sprintf("num leaves with %s version = %v", core.GetStringForVersion(1), 2), trieStatsStrings[11])
	assert.Equal(t, fmt.Sprintf("num leaves with %s version = %v", core.GetStringForVersion(0), 2), trieStatsStrings[12])
}
