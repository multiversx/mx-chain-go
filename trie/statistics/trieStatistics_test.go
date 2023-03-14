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
