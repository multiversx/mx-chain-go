package staking

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func requireSameSliceDifferentOrder(t *testing.T, s1, s2 [][]byte) {
	require.Equal(t, len(s1), len(s2))

	for _, elemInS1 := range s1 {
		require.Contains(t, s2, elemInS1)
	}
}

func searchInMap(validatorMap map[uint32][][]byte, pk []byte) bool {
	for _, validatorsInShard := range validatorMap {
		for _, val := range validatorsInShard {
			if bytes.Equal(val, pk) {
				return true
			}
		}
	}
	return false
}

func requireMapContains(t *testing.T, m map[uint32][][]byte, s [][]byte) {
	for _, elemInSlice := range s {
		require.True(t, searchInMap(m, elemInSlice))
	}

}

func getAllPubKeys(validatorsMap map[uint32][][]byte) [][]byte {
	allValidators := make([][]byte, 0)
	for _, validatorsInShard := range validatorsMap {
		allValidators = append(allValidators, validatorsInShard...)
	}

	return allValidators
}

func TestNewTestMetaProcessor(t *testing.T) {
	numOfMetaNodes := uint32(10)
	numOfShards := uint32(3)
	numOfEligibleNodesPerShard := uint32(10)
	numOfWaitingNodesPerShard := uint32(10)
	numOfNodesToShufflePerShard := uint32(3)
	shardConsensusGroupSize := 3
	metaConsensusGroupSize := 3
	numOfNodesInStakingQueue := uint32(4)

	totalEligible := int(numOfEligibleNodesPerShard*numOfShards) + int(numOfMetaNodes)
	totalWaiting := int(numOfWaitingNodesPerShard*numOfShards) + int(numOfMetaNodes)

	node := NewTestMetaProcessor(
		numOfMetaNodes,
		numOfShards,
		numOfEligibleNodesPerShard,
		numOfWaitingNodesPerShard,
		numOfNodesToShufflePerShard,
		shardConsensusGroupSize,
		metaConsensusGroupSize,
		numOfNodesInStakingQueue,
	)
	node.EpochStartTrigger.SetRoundsPerEpoch(4)

	initialNodes := node.NodesConfig
	require.Len(t, getAllPubKeys(initialNodes.eligible), totalEligible)
	require.Len(t, getAllPubKeys(initialNodes.waiting), totalWaiting)
	require.Len(t, initialNodes.queue, int(numOfNodesInStakingQueue))
	require.Empty(t, initialNodes.shuffledOut)
	require.Empty(t, initialNodes.auction)

	node.Process(t, 5)
	nodesConfigStakingV4Init := node.NodesConfig
	require.Len(t, getAllPubKeys(nodesConfigStakingV4Init.eligible), totalEligible)
	require.Len(t, getAllPubKeys(nodesConfigStakingV4Init.waiting), totalWaiting)
	require.Empty(t, nodesConfigStakingV4Init.queue)
	require.Empty(t, nodesConfigStakingV4Init.shuffledOut)
	requireSameSliceDifferentOrder(t, initialNodes.queue, nodesConfigStakingV4Init.auction)

	node.Process(t, 6)
	nodesConfigStakingV4 := node.NodesConfig
	require.Len(t, getAllPubKeys(nodesConfigStakingV4.eligible), totalEligible)
	require.Len(t, getAllPubKeys(nodesConfigStakingV4.waiting), totalWaiting-int((numOfShards+1)*numOfNodesToShufflePerShard)+len(nodesConfigStakingV4Init.auction))

	requireMapContains(t, nodesConfigStakingV4.waiting, nodesConfigStakingV4Init.auction)  // all current waiting are from the previous auction
	requireMapContains(t, nodesConfigStakingV4Init.eligible, nodesConfigStakingV4.auction) // all current auction are from previous eligible

	//requireMapContains(t, node.NodesConfig.shuffledOut, node.NodesConfig.auction, uint32(len(node.NodesConfig.shuffledOut)))
	//requireMapContains(t, eligibleAfterStakingV4Init, node.NodesConfig.auction, 8) //todo: check size

	//node.Process(t, 20)
}
