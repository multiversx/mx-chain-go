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
	numOfMetaNodes := uint32(400)
	numOfShards := uint32(3)
	numOfEligibleNodesPerShard := uint32(400)
	numOfWaitingNodesPerShard := uint32(400)
	numOfNodesToShufflePerShard := uint32(80)
	shardConsensusGroupSize := 266
	metaConsensusGroupSize := 266
	numOfNodesInStakingQueue := uint32(60)

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

	// 1. Check initial config is correct
	initialNodes := node.NodesConfig
	require.Len(t, getAllPubKeys(initialNodes.eligible), totalEligible)
	require.Len(t, getAllPubKeys(initialNodes.waiting), totalWaiting)
	require.Len(t, initialNodes.queue, int(numOfNodesInStakingQueue))
	require.Empty(t, initialNodes.shuffledOut)
	require.Empty(t, initialNodes.auction)

	// 2. Check config after staking v4 initialization
	node.Process(t, 5)
	nodesConfigStakingV4Init := node.NodesConfig
	require.Len(t, getAllPubKeys(nodesConfigStakingV4Init.eligible), totalEligible)
	require.Len(t, getAllPubKeys(nodesConfigStakingV4Init.waiting), totalWaiting)
	require.Empty(t, nodesConfigStakingV4Init.queue)
	require.Empty(t, nodesConfigStakingV4Init.shuffledOut)
	requireSameSliceDifferentOrder(t, initialNodes.queue, nodesConfigStakingV4Init.auction)

	// 3. Check config after first staking v4 epoch
	node.Process(t, 6)
	nodesConfigStakingV4 := node.NodesConfig
	require.Len(t, getAllPubKeys(nodesConfigStakingV4.eligible), totalEligible)

	numOfShuffledOut := int((numOfShards + 1) * numOfNodesToShufflePerShard)
	newWaiting := totalWaiting - numOfShuffledOut + len(nodesConfigStakingV4Init.auction)
	require.Len(t, getAllPubKeys(nodesConfigStakingV4.waiting), newWaiting)

	// All shuffled out are in auction
	require.Len(t, getAllPubKeys(nodesConfigStakingV4.shuffledOut), numOfShuffledOut)
	requireSameSliceDifferentOrder(t, getAllPubKeys(nodesConfigStakingV4.shuffledOut), nodesConfigStakingV4.auction)

	// All current waiting are from the previous auction
	requireMapContains(t, nodesConfigStakingV4.waiting, nodesConfigStakingV4Init.auction)
	// All current auction are from previous eligible
	requireMapContains(t, nodesConfigStakingV4Init.eligible, nodesConfigStakingV4.auction)

	epochs := 0
	prevConfig := nodesConfigStakingV4
	prevNumOfWaiting := newWaiting
	for epochs < 10 {
		node.Process(t, 5)
		newNodeConfig := node.NodesConfig

		newWaiting = prevNumOfWaiting - numOfShuffledOut + len(prevConfig.auction)
		require.Len(t, getAllPubKeys(newNodeConfig.waiting), newWaiting)
		require.Len(t, getAllPubKeys(newNodeConfig.eligible), totalEligible)

		require.Len(t, getAllPubKeys(newNodeConfig.shuffledOut), numOfShuffledOut)
		requireSameSliceDifferentOrder(t, getAllPubKeys(newNodeConfig.shuffledOut), newNodeConfig.auction)

		requireMapContains(t, newNodeConfig.waiting, prevConfig.auction)
		requireMapContains(t, prevConfig.eligible, newNodeConfig.auction)

		prevConfig = newNodeConfig
		prevNumOfWaiting = newWaiting
		epochs++
	}
}
